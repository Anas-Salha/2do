//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const basePath = "/api/v0/todos"

type Todo struct {
	ID        uint32    `json:"id"`
	Todo      string    `json:"todo"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TodoInput struct {
	Todo      *string `json:"todo"`
	Completed *bool   `json:"completed"`
}

func baseURL() string {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://0.0.0.0:8080"
}

func waitReady(t *testing.T, base string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/healthz")
		if err == nil {
			if resp.StatusCode == 204 {
				return
			}
			lastErr = fmt.Errorf("status=%d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}
	require.FailNowf(t, "service did not become ready", "base=%s err=%v", base, lastErr)
}

func mustJSON[T any](t *testing.T, v T) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func doReq(t *testing.T, method, url string, body any, hdr map[string]string) *http.Response {
	t.Helper()
	var r io.Reader
	switch b := body.(type) {
	case nil:
	case []byte:
		r = bytes.NewReader(b)
	case string:
		r = strings.NewReader(b)
	default:
		r = bytes.NewReader(mustJSON(t, b))
		if hdr == nil {
			hdr = map[string]string{}
		}
		if _, ok := hdr["Content-Type"]; !ok {
			hdr["Content-Type"] = "application/json"
		}
	}
	req, err := http.NewRequest(method, url, r)
	require.NoError(t, err)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decode[T any](t *testing.T, rc io.ReadCloser) T {
	t.Helper()
	defer rc.Close()
	var v T
	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoErrorf(t, json.Unmarshal(b, &v), "decode failed; body=%s", string(b))
	return v
}

func randTitle(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz "
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		s = "x"
	}
	return strings.Title(s)
}

func envInt(name string, def int) int {
	if s := os.Getenv(name); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractIDFromLocation(t *testing.T, loc string) uint32 {
	t.Helper()

	// Strip scheme/host if present
	if u, err := url.Parse(loc); err == nil && u != nil && u.Path != "" {
		loc = u.Path
	}

	parts := strings.Split(strings.Trim(loc, "/"), "/")
	require.GreaterOrEqual(t, len(parts), 1, "invalid Location: %q", loc)

	last := parts[len(parts)-1]
	n, err := strconv.ParseUint(last, 10, 32)
	require.NoError(t, err, "invalid id in Location: %q", loc)

	return uint32(n)
}

// -------------------- TESTS --------------------

func TestE2E_FailCases(t *testing.T) {
	base := baseURL()
	waitReady(t, base, 10*time.Second)

	t.Run("GET_nonexistent", func(t *testing.T) {
		resp := doReq(t, http.MethodGet, base+basePath+"/99999999", nil, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("GET_bad_id", func(t *testing.T) {
		resp := doReq(t, http.MethodGet, base+basePath+"/abc", nil, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("POST_empty_body", func(t *testing.T) {
		resp := doReq(t, http.MethodPost, base+basePath, "", map[string]string{"Content-Type": "application/json"})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("POST_invalid_json", func(t *testing.T) {
		resp := doReq(t, http.MethodPost, base+basePath, "{not-json", map[string]string{"Content-Type": "application/json"})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("POST_missing_required_fields", func(t *testing.T) {
		in := TodoInput{} // nil fields -> should fail create
		resp := doReq(t, http.MethodPost, base+basePath, in, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("PUT_nonexistent", func(t *testing.T) {
		msg := "x"
		in := TodoInput{Todo: &msg}
		resp := doReq(t, http.MethodPut, base+basePath+"/1234567", in, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("PUT_invalid_json", func(t *testing.T) {
		resp := doReq(t, http.MethodPut, base+basePath+"/1", "{invalid", map[string]string{"Content-Type": "application/json"})
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("PUT_valid_json_empty_object", func(t *testing.T) {
		// seed a todo to ensure ID exists
		title := "seed-empty-object"
		comp := false
		in := TodoInput{Todo: &title, Completed: &comp}
		resp := doReq(t, http.MethodPost, base+basePath, in, nil)
		require.Contains(t, []int{http.StatusCreated, http.StatusOK}, resp.StatusCode)

		loc := resp.Header.Get("Location")
		require.NotEmpty(t, loc, "Location header must be set on create")
		resp.Body.Close()

		id := extractIDFromLocation(t, loc)

		// valid JSON but empty object {}
		resp = doReq(t, http.MethodPut, fmt.Sprintf("%s%s/%d", base, basePath, id), map[string]any{}, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("PUT_valid_json_null_fields", func(t *testing.T) {
		title := "seed-nulls"
		comp := false
		in := TodoInput{Todo: &title, Completed: &comp}
		resp := doReq(t, http.MethodPost, base+basePath, in, nil)
		require.Contains(t, []int{http.StatusCreated, http.StatusOK}, resp.StatusCode)

		loc := resp.Header.Get("Location")
		require.NotEmpty(t, loc, "Location header must be set on create")
		resp.Body.Close()

		id := extractIDFromLocation(t, loc)

		// valid JSON with explicit nulls
		resp = doReq(t, http.MethodPut, fmt.Sprintf("%s%s/%d", base, basePath, id),
			map[string]any{"todo": nil, "completed": nil}, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("PUT_id_zero_with_valid_json", func(t *testing.T) {
		// id=0 should not be valid, even with a proper body
		title := "any"
		up := TodoInput{Todo: &title}
		resp := doReq(t, http.MethodPut, base+basePath+"/0", up, nil)
		require.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("METHOD_not_allowed", func(t *testing.T) {
		resp := doReq(t, http.MethodPatch, base+basePath, map[string]any{"todo": "nope"}, nil)
		require.Contains(t, []int{http.StatusMethodNotAllowed, http.StatusNotFound}, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestE2E_PartialUpdate_Steps(t *testing.T) {
	base := baseURL()
	waitReady(t, base, 10*time.Second)

	// Seed a todo
	title := "seed-partial"
	comp := false
	in := TodoInput{Todo: &title, Completed: &comp}
	resp := doReq(t, http.MethodPost, base+basePath, in, nil)
	require.Contains(t, []int{http.StatusCreated, http.StatusOK}, resp.StatusCode)

	loc := resp.Header.Get("Location")
	require.NotEmpty(t, loc, "Location header must be set on create")
	resp.Body.Close()

	id := extractIDFromLocation(t, loc)

	t.Run("PUT_update_completed_only", func(t *testing.T) {
		newCompleted := true
		resp := doReq(t, http.MethodPut,
			fmt.Sprintf("%s%s/%d", base, basePath, id),
			TodoInput{Completed: &newCompleted}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// Verify with GET
		resp = doReq(t, http.MethodGet, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
		got := decode[Todo](t, resp.Body)
		require.Equal(t, title, got.Todo, "title should remain unchanged")
		require.True(t, got.Completed, "completed should be updated to true")
	})

	t.Run("PUT_update_todo_only", func(t *testing.T) {
		newTitle := "seed-partial-updated"
		resp := doReq(t, http.MethodPut,
			fmt.Sprintf("%s%s/%d", base, basePath, id),
			TodoInput{Todo: &newTitle}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// Verify with GET
		resp = doReq(t, http.MethodGet, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
		got := decode[Todo](t, resp.Body)
		require.Equal(t, newTitle, got.Todo, "todo should be updated")
		require.True(t, got.Completed, "completed should remain unchanged")
	})
}

func TestE2E_CoreCRUD_AndLightSoak(t *testing.T) {
	base := baseURL()
	waitReady(t, base, 10*time.Second)

	// CREATE (POST) -> 201 + Location header (no entity in body)
	title := "Write E2E"
	comp := false
	in := TodoInput{Todo: &title, Completed: &comp}
	resp := doReq(t, http.MethodPost, base+basePath, in, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	loc := resp.Header.Get("Location")
	require.NotEmpty(t, loc, "Location header must be set on create")
	resp.Body.Close()

	id := extractIDFromLocation(t, loc)

	// GET created entity to confirm fields/timestamps
	resp = doReq(t, http.MethodGet, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
	created := decode[Todo](t, resp.Body)
	require.Equal(t, id, created.ID)
	require.Equal(t, title, created.Todo)
	require.Equal(t, comp, created.Completed)
	require.False(t, created.CreatedAt.IsZero())
	require.False(t, created.UpdatedAt.IsZero())
	require.False(t, created.UpdatedAt.Before(created.CreatedAt))

	// UPDATE (PUT) -> 200 {"updated": true} (no entity) then GET to verify
	time.Sleep(1 * time.Second) // to ensure timestamp change
	newComp := true
	up := TodoInput{Completed: &newComp}
	resp = doReq(t, http.MethodPut, fmt.Sprintf("%s%s/%d", base, basePath, id), up, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify only 'completed' changed, title unchanged
	resp = doReq(t, http.MethodGet, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
	updated := decode[Todo](t, resp.Body)
	require.Equal(t, id, updated.ID)
	require.Equal(t, title, updated.Todo) // unchanged
	require.True(t, updated.Completed)    // updated
	require.NotEqual(t, created.UpdatedAt, updated.UpdatedAt)
	require.False(t, created.UpdatedAt.Before(created.CreatedAt))

	// LIST (no pagination yet)
	resp = doReq(t, http.MethodGet, base+basePath, nil, nil)
	all := decode[[]Todo](t, resp.Body)
	require.GreaterOrEqual(t, len(all), 1)

	// LIGHT SOAK: repeated list for ~5s
	deadline := time.Now().Add(5 * time.Second)
	n := 0
	for time.Now().Before(deadline) {
		resp = doReq(t, http.MethodGet, base+basePath, nil, nil)
		_ = decode[[]Todo](t, resp.Body)
		n++
		time.Sleep(100 * time.Millisecond)
	}
	require.Greater(t, n, 20)

	// DELETE -> 204, then DELETE again -> 404
	resp = doReq(t, http.MethodDelete, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	resp = doReq(t, http.MethodDelete, fmt.Sprintf("%s%s/%d", base, basePath, id), nil, nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

// ---- STRESS (skipped unless E2E_STRESS=1) ----

func TestE2E_Stress_BulkCreate_List_ConcurrentPuts(t *testing.T) {
	if os.Getenv("E2E_STRESS") != "1" {
		t.Skip("stress suite skipped (set E2E_STRESS=1 to enable)")
	}

	base := baseURL()
	waitReady(t, base, 10*time.Second)

	total := envInt("E2E_TODOS", 3000)
	workers := envInt("E2E_WORKERS", 24)

	created := make([]uint32, 0, total)

	// BULK CREATE
	t.Run("bulk_create", func(t *testing.T) {
		type res struct {
			id  uint32
			err error
		}
		jobs := make(chan int, total)
		out := make(chan res, total)

		for w := 0; w < workers; w++ {
			go func() {
				for i := range jobs {
					title := randTitle(14)
					comp := (i%5 == 0)
					in := TodoInput{Todo: &title, Completed: &comp}

					resp := doReq(t, http.MethodPost, base+basePath, in, nil)
					if resp.StatusCode != http.StatusCreated {
						// On error, body may contain details
						b, _ := io.ReadAll(resp.Body)
						resp.Body.Close()
						out <- res{0, fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b))}
						continue
					}

					loc := resp.Header.Get("Location")
					resp.Body.Close()
					if loc == "" {
						out <- res{0, fmt.Errorf("missing Location header")}
						continue
					}
					id := extractIDFromLocation(t, loc)
					out <- res{id, nil}
				}
			}()
		}

		for i := 0; i < total; i++ {
			jobs <- i
		}
		close(jobs)

		for i := 0; i < total; i++ {
			r := <-out
			require.NoError(t, r.err)
			require.NotZero(t, r.id)
			created = append(created, r.id)
		}
	})

	// LIST (sanity)
	t.Run("list_all", func(t *testing.T) {
		resp := doReq(t, http.MethodGet, base+basePath, nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		all := decode[[]Todo](t, resp.Body)
		require.GreaterOrEqual(t, len(all), len(created))
		// spot-check timestamps
		for _, td := range all[:min(5, len(all))] {
			require.False(t, td.CreatedAt.IsZero())
			require.False(t, td.UpdatedAt.IsZero())
			require.True(t, !td.UpdatedAt.Before(td.CreatedAt))
		}
	})

	// CONCURRENT PUTs against the same row
	t.Run("concurrent_puts", func(t *testing.T) {
		title := "base"
		f := false
		resp := doReq(t, http.MethodPost, base+basePath, TodoInput{Todo: &title, Completed: &f}, nil)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		loc := resp.Header.Get("Location")
		require.NotEmpty(t, loc, "Location header required on create")
		resp.Body.Close()
		rowID := extractIDFromLocation(t, loc)

		n := 32
		var wg sync.WaitGroup
		wg.Add(n)
		errs := make(chan error, n)

		for i := 0; i < n; i++ {
			go func(i int) {
				defer wg.Done()
				todo := fmt.Sprintf("v%d", i)
				comp := i%2 == 0
				in := TodoInput{Todo: &todo, Completed: &comp}

				resp := doReq(t, http.MethodPut, fmt.Sprintf("%s%s/%d", base, basePath, rowID), in, nil)
				if resp.StatusCode != http.StatusOK {
					// On error, body may contain details
					b, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					errs <- fmt.Errorf("status=%d body=%s", resp.StatusCode, string(b))
					return
				}
				// No body expected on success
				resp.Body.Close()
			}(i)
		}
		wg.Wait()
		close(errs)
		for e := range errs {
			require.NoError(t, e)
		}

		// Verify the row still exists and is retrievable
		resp = doReq(t, http.MethodGet, fmt.Sprintf("%s%s/%d", base, basePath, rowID), nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_ = decode[Todo](t, resp.Body)
	})
}
