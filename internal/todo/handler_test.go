package todo_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/anas-salha/2do/internal/todo"
)

type mockService struct {
	getAllFn  func(context.Context) ([]Todo, error)
	getByIDFn func(context.Context, uint32) (*Todo, error)
	createFn  func(context.Context, TodoInput) (*Todo, error)
	updateFn  func(context.Context, uint32, TodoInput) (*Todo, error)
	deleteFn  func(context.Context, uint32) error
}

var _ Service = (*mockService)(nil)

func (m *mockService) GetAll(ctx context.Context) ([]Todo, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockService) GetById(ctx context.Context, id uint32) (*Todo, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockService) Create(ctx context.Context, in TodoInput) (*Todo, error) {
	if m.createFn != nil {
		return m.createFn(ctx, in)
	}
	return nil, nil
}

func (m *mockService) Update(ctx context.Context, id uint32, in TodoInput) (*Todo, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, in)
	}
	return nil, nil
}

func (m *mockService) Delete(ctx context.Context, id uint32) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

var _ = Describe("handler", Label("handler"), func() {
	var (
		svc    *mockService
		router *gin.Engine
		rr     *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		svc = &mockService{}
		router = gin.New()
		NewHandler(svc).Register(router)
		rr = httptest.NewRecorder()
	})

	Describe("GET /todos", Label("get-all"), func() {
		It("Verifies happy path", func() {
			expected := []Todo{{ID: 1, Text: "walk the dog"}, {ID: 2, Text: "feed the cat"}}
			svc.getAllFn = func(ctx context.Context) ([]Todo, error) { return expected, nil }

			req := httptest.NewRequest(http.MethodGet, "/todos", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))

			var out []Todo
			Expect(json.Unmarshal(rr.Body.Bytes(), &out)).To(Succeed())
			Expect(out).To(Equal(expected))
		})

		It("Reports internal server error", func() {
			err := errors.New("database is down")
			svc.getAllFn = func(ctx context.Context) ([]Todo, error) { return nil, err }

			req := httptest.NewRequest(http.MethodGet, "/todos", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"unexpected error"}`))
		})
	})

	Describe("GET /todos/:id", Label("get-id"), func() {
		It("Verifies happy path", func() {
			svc.getByIDFn = func(ctx context.Context, id uint32) (*Todo, error) {
				Expect(id).To(Equal(uint32(3)))
				return &Todo{ID: id, Text: "buy groceries"}, nil
			}

			req := httptest.NewRequest(http.MethodGet, "/todos/3", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))

			var out Todo
			Expect(json.Unmarshal(rr.Body.Bytes(), &out)).To(Succeed())
			Expect(out.ID).To(Equal(uint32(3)))
		})

		It("Reports bad request for invalid ID type", func() {
			req := httptest.NewRequest(http.MethodGet, "/todos/x", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"ID must be an integer"}`))
		})

		It("Propagates error not found when ID doesn't exist", func() {
			svc.getByIDFn = func(ctx context.Context, id uint32) (*Todo, error) { return nil, ErrNotFound }

			req := httptest.NewRequest(http.MethodGet, "/todos/3", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusNotFound))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"todo not found"}`))
		})

		It("Reports internal server error", func() {
			err := errors.New("database is down")
			svc.getByIDFn = func(ctx context.Context, id uint32) (*Todo, error) { return nil, err }

			req := httptest.NewRequest(http.MethodGet, "/todos/3", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"unexpected error"}`))
		})
	})

	Describe("POST /todos", Label("post"), func() {
		It("Verifies happy path - default completed status", func() {
			svc.createFn = func(ctx context.Context, in TodoInput) (*Todo, error) {
				Expect(in.Text).NotTo(BeNil())
				Expect(*in.Text).To(Equal("stretch"))
				Expect(in.Completed).NotTo(BeNil())
				Expect(*in.Completed).To(BeFalse())
				return &Todo{ID: 7, Text: "stretch", Completed: false}, nil
			}

			payload := "{\"text\":\"stretch\"}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusCreated))
			Expect(rr.Header().Get("Location")).To(Equal("/todos/7"))
			Expect(rr.Body.String()).To(ContainSubstring(`"id":7,"text":"stretch","completed":false`))
		})

		It("Verifies happy path - provided completed status", func() {
			svc.createFn = func(ctx context.Context, in TodoInput) (*Todo, error) {
				Expect(in.Text).NotTo(BeNil())
				Expect(*in.Text).To(Equal("stretch"))
				Expect(in.Completed).NotTo(BeNil())
				Expect(*in.Completed).To(BeTrue())
				return &Todo{ID: 7, Text: "stretch", Completed: true}, nil
			}

			payload := "{\"text\":\"stretch\",\"completed\":true}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusCreated))
			Expect(rr.Header().Get("Location")).To(Equal("/todos/7"))
			Expect(rr.Body.String()).To(ContainSubstring(`"id":7,"text":"stretch","completed":true`))
		})

		It("Reports bad request for invalid JSON", func() {
			payload := "{\"text\":true}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"invalid json"}`))
		})

		It("Reports error unsupported media type", func() {
			payload := "{\"text\":\"water the plants\"}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/octal-stream")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusUnsupportedMediaType))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"Content-Type must be application/json"}`))
		})

		It("Propagates error input invalid", func() {
			svc.createFn = func(ctx context.Context, in TodoInput) (*Todo, error) { return nil, ErrInputInvalid }

			payload := "{\"text\":\"\"}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"todo input invalid"}`))
		})

		It("Reports internal server error", func() {
			err := errors.New("database is down")
			svc.createFn = func(ctx context.Context, in TodoInput) (*Todo, error) { return nil, err }

			payload := "{\"text\":\"stretch\"}"
			req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"unexpected error"}`))
		})
	})

	Describe("PUT /todos", Label("put"), func() {
		It("Verifies happy path", func() {
			svc.updateFn = func(ctx context.Context, id uint32, in TodoInput) (*Todo, error) {
				Expect(id).To(Equal(uint32(12)))
				Expect(in.Text).NotTo(BeNil())
				Expect(*in.Text).To(Equal("call mom"))
				Expect(in.Completed).NotTo(BeNil())
				Expect(*in.Completed).To(BeTrue())
				return &Todo{ID: id, Text: "call mom", Completed: true}, nil
			}

			payload := "{\"text\":\"call mom\",\"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/12", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
		})

		It("Reports bad request for invalid ID type", func() {
			req := httptest.NewRequest(http.MethodPut, "/todos/x", nil)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"ID must be an integer"}`))
		})

		It("Reports bad request for invalid JSON", func() {
			payload := "{\"text\":true,\"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"invalid json"}`))
		})

		It("Reports bad request for missing field", func() {
			payload := "{\"text\":\"call mom\"}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"missing field"}`))
		})

		It("Reports error unsupported media type", func() {
			payload := "{\"text\":\"water the plants\",\"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/octal-stream")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusUnsupportedMediaType))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"Content-Type must be application/json"}`))
		})

		It("Propagates error not found when ID doesn't exist", func() {
			svc.updateFn = func(ctx context.Context, id uint32, in TodoInput) (*Todo, error) { return nil, ErrNotFound }
			payload := "{\"text\":\"call mom\",\"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusNotFound))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"todo not found"}`))
		})

		It("Propagates error input invalid", func() {
			svc.updateFn = func(ctx context.Context, id uint32, in TodoInput) (*Todo, error) { return nil, ErrInputInvalid }
			payload := "{\"text\":\"\", \"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusUnprocessableEntity))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"todo input invalid"}`))
		})

		It("Reports internal server error", func() {
			err := errors.New("database is down")
			svc.updateFn = func(ctx context.Context, id uint32, in TodoInput) (*Todo, error) { return nil, err }

			payload := "{\"text\":\"call mom\",\"completed\":true}"
			req := httptest.NewRequest(http.MethodPut, "/todos/3", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"unexpected error"}`))
		})
	})

	Describe("DELETE /todos/:id", Label("delete"), func() {
		It("Verifies happy path", func() {
			svc.deleteFn = func(ctx context.Context, id uint32) error {
				Expect(id).To(Equal(uint32(5)))
				return nil
			}

			req := httptest.NewRequest(http.MethodDelete, "/todos/5", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusNoContent))
		})

		It("Reports bad request for invalid ID type", func() {
			req := httptest.NewRequest(http.MethodDelete, "/todos/x", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"ID must be an integer"}`))
		})

		It("Propagates error not found when ID doesn't exist", func() {
			svc.deleteFn = func(ctx context.Context, id uint32) error { return ErrNotFound }
			req := httptest.NewRequest(http.MethodDelete, "/todos/3", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusNotFound))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"todo not found"}`))
		})

		It("Reports internal server error", func() {
			err := errors.New("database is down")
			svc.deleteFn = func(ctx context.Context, id uint32) error { return err }
			req := httptest.NewRequest(http.MethodDelete, "/todos/3", nil)
			router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(BeEquivalentTo(`{"error":"unexpected error"}`))
		})
	})
})
