package todo_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/anas-salha/2do/internal/todo"
)

var _ = Describe("repo", Label("repo"), func() {
	var (
		ctx  context.Context
		db   *sql.DB
		mock sqlmock.Sqlmock
		repo Repository
		now  time.Time
		rows *sqlmock.Rows
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		db, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		repo = NewRepo(db)
		now = time.Now().UTC().Truncate(time.Second)
		rows = sqlmock.NewRows([]string{"id", "text", "completed", "created_at", "updated_at"})
	})

	AfterEach(func() {
		mock.ExpectClose()
		Expect(db.Close()).To(Succeed())
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Describe("List", Label("list"), func() {
		var query string

		BeforeEach(func() {
			query = "SELECT id, text, completed, created_at, updated_at FROM `todos`"
		})

		It("lists no todos (empty) successfully", func() {
			mock.ExpectQuery(query).WillReturnRows(rows)

			todos, err := repo.List(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(todos).To(BeEmpty())
		})

		It("lists 2 todos successfully", func() {
			rows = rows.
				AddRow(1, "walk the dog", false, now, now).
				AddRow(2, "buy groceries", true, now, now)
			mock.ExpectQuery(query).WillReturnRows(rows)

			todos, err := repo.List(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(todos).To(HaveLen(2))
			Expect(todos[0].Text).To(Equal("walk the dog"))
			Expect(todos[0].Completed).To(BeFalse())
			Expect(todos[1].Text).To(Equal("buy groceries"))
			Expect(todos[1].Completed).To(BeTrue())
		})

		It("propagates query errors", func() {
			expected := errors.New("list failed")
			mock.ExpectQuery(query).WillReturnError(expected)

			todos, err := repo.List(ctx)
			Expect(err).To(MatchError(expected))
			Expect(todos).To(BeNil())
		})

		It("propagates scan errors", func() {
			rows = rows.
				AddRow(1, "walk the dog", false, now, now).
				AddRow(2, nil, true, now, now) // text is nil, should cause scan error
			mock.ExpectQuery(query).WillReturnRows(rows)

			todos, err := repo.List(ctx)
			Expect(err).To(HaveOccurred())
			Expect(todos).To(BeNil())
		})

		It("propagates row iteration errors", func() {
			rows = rows.
				AddRow(1, "walk the dog", false, now, now).
				RowError(0, errors.New("row iteration error")).
				AddRow(2, "buy groceries", true, now, now)

			mock.ExpectQuery(query).WillReturnRows(rows)

			todos, err := repo.List(ctx)
			Expect(err).To(MatchError("row iteration error"))
			Expect(todos).To(BeNil())
		})
	})

	Describe("Get", Label("get"), func() {
		var query string

		BeforeEach(func() {
			query = "SELECT id, text, completed, created_at, updated_at FROM `todos` WHERE id=?"
		})

		It("get todo successfully", func() {
			id := uint32(1)
			rows = rows.
				AddRow(id, "walk the dog", false, now, now)
			mock.ExpectQuery(query).WillReturnRows(rows)

			todo, err := repo.Get(ctx, id)
			Expect(err).NotTo(HaveOccurred())
			Expect(todo.Text).To(Equal("walk the dog"))
			Expect(todo.Completed).To(BeFalse())
		})

		It("propagates query errors", func() {
			expected := errors.New("get failed")
			mock.ExpectQuery(query).WillReturnError(expected)

			todo, err := repo.Get(ctx, 1)
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})

		It("returns todo not found errors", func() {
			mock.ExpectQuery(query).WillReturnError(sql.ErrNoRows)

			todo, err := repo.Get(ctx, 1)
			Expect(err).To(MatchError(ErrNotFound))
			Expect(todo).To(BeNil())
		})
	})

	Describe("Create", Label("create"), func() {
		var (
			query    string
			getQuery string
		)

		BeforeEach(func() {
			query = "INSERT INTO `todos` (text) VALUES ('%s')"
			getQuery = "SELECT id, text, completed, created_at, updated_at FROM `todos` WHERE id=?"
		})

		It("creates and returns a todo successfully", func() {
			text := "hit the gym"
			input := TodoInput{Text: &text}

			query = fmt.Sprintf(query, text)
			lastInsertId := int64(3)
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(lastInsertId, 1))

			rows = rows.AddRow(lastInsertId, text, false, now, now)
			mock.ExpectQuery(getQuery).WillReturnRows(rows)

			todo, err := repo.Create(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(todo).NotTo(BeNil())
			Expect(todo.ID).To(BeEquivalentTo(lastInsertId))
			Expect(todo.Text).To(Equal(text))
			Expect(todo.Completed).To(BeFalse())
		})

		It("propagates insert errors", func() {
			text := "dummy todo"
			input := TodoInput{Text: &text}

			query = fmt.Sprintf(query, text)
			expected := errors.New("insert failed")
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(expected)

			todo, err := repo.Create(ctx, input)
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})

		It("propagates error from LastInsertId", func() {
			text := "dummy todo"
			input := TodoInput{Text: &text}

			query = fmt.Sprintf(query, text)
			expected := errors.New("lastInsertId failed")
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewErrorResult(expected))

			todo, err := repo.Create(ctx, input)
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})

		It("propagates get errors", func() {
			text := "dummy todo"
			input := TodoInput{Text: &text}

			query = fmt.Sprintf(query, text)
			lastInsertId := int64(3)
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(lastInsertId, 1))

			expected := errors.New("get failed")
			mock.ExpectQuery(getQuery).WillReturnError(expected)

			todo, err := repo.Create(ctx, input)
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})
	})

	Describe("Update", Label("update"), func() {
		var (
			query    string
			getQuery string
		)

		BeforeEach(func() {
			query = "UPDATE `todos` SET text = IFNULL(?, text), completed = IFNULL(?, completed) WHERE id=?"
			getQuery = "SELECT id, text, completed, created_at, updated_at FROM `todos` WHERE id=?"
		})

		It("updates only text and returns a todo successfully", func() {
			id := 3
			text := "hit the gym"
			input := TodoInput{Text: &text}

			mock.ExpectExec(regexp.QuoteMeta(query)).
				WithArgs(&text, nil, id).
				WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1

			rows = rows.AddRow(id, text, false, now, now)
			mock.ExpectQuery(getQuery).WillReturnRows(rows)

			todo, err := repo.Update(ctx, uint32(id), input)
			Expect(err).NotTo(HaveOccurred())
			Expect(todo).NotTo(BeNil())
			Expect(todo.Text).To(Equal(text))
			Expect(todo.Completed).To(BeFalse())
		})

		It("updates only completed and returns a todo successfully", func() {
			id := 3
			completed := true
			input := TodoInput{Completed: &completed}
			text := "dummy todo"

			mock.ExpectExec(regexp.QuoteMeta(query)).
				WithArgs(nil, &completed, id).
				WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1

			rows = rows.AddRow(id, text, &completed, now, now)
			mock.ExpectQuery(getQuery).WillReturnRows(rows)

			todo, err := repo.Update(ctx, uint32(id), input)
			Expect(err).NotTo(HaveOccurred())
			Expect(todo).NotTo(BeNil())
			Expect(todo.Text).To(Equal(text))
			Expect(todo.Completed).To(BeTrue())
		})

		It("updates text & completed and returns a todo successfully", func() {
			id := 3
			completed := true
			text := "make dinner"
			input := TodoInput{Text: &text, Completed: &completed}

			mock.ExpectExec(regexp.QuoteMeta(query)).
				WithArgs(&text, &completed, id).
				WillReturnResult(sqlmock.NewResult(0, 1)) // RowsAffected = 1

			rows = rows.AddRow(id, text, &completed, now, now)
			mock.ExpectQuery(getQuery).WillReturnRows(rows)

			todo, err := repo.Update(ctx, uint32(id), input)
			Expect(err).NotTo(HaveOccurred())
			Expect(todo).NotTo(BeNil())
			Expect(todo.Text).To(Equal(text))
			Expect(todo.Completed).To(BeTrue())
		})

		It("propagates update errors", func() {
			expected := errors.New("update failed")
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(expected)

			todo, err := repo.Update(ctx, 1, TodoInput{})
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})

		It("propagates get errors", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 1)) // Mock success inserting

			expected := errors.New("get failed")
			mock.ExpectQuery(getQuery).WillReturnError(expected)

			todo, err := repo.Update(ctx, 1, TodoInput{})
			Expect(err).To(MatchError(expected))
			Expect(todo).To(BeNil())
		})

		It("returns todo not found error if no rows affected", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 0))

			todo, err := repo.Update(ctx, 1, TodoInput{})
			Expect(err).To(MatchError(ErrNotFound))
			Expect(todo).To(BeNil())
		})

		It("propagates error if multiple rows affected", func() {
			mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(0, 3))

			todo, err := repo.Update(ctx, 1, TodoInput{})
			Expect(err).To(MatchError(ErrMultipleRowsAffected))
			Expect(todo).To(BeNil())
		})

	})

	Describe("Delete", Label("delete"), func() {
		var query string

		BeforeEach(func() {
			query = "DELETE FROM `todos` WHERE id=?"
		})

		It("deletes todo successfully", func() {
			mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.Delete(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
		})

		It("propagates delete errors", func() {
			expected := errors.New("insert failed")
			mock.ExpectExec(query).WillReturnError(expected)

			err := repo.Delete(ctx, 1)
			Expect(err).To(MatchError(expected))
		})

		It("returns todo not found error if no rows affected", func() {
			mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.Delete(ctx, 1)
			Expect(err).To(MatchError(ErrNotFound))
		})

		It("propagates error if multiple rows affected", func() {
			mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 3))

			err := repo.Delete(ctx, 1)
			Expect(err).To(MatchError(ErrMultipleRowsAffected))
		})

	})
})
