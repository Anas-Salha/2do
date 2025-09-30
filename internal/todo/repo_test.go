package todo_test

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/anas-salha/2do/internal/todo"
)

var _ = Describe("repo", func() {
	var (
		ctx  context.Context
		db   *sql.DB
		mock sqlmock.Sqlmock
		repo todo.Repository
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		db, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		repo = todo.NewRepo(db)
	})

	AfterEach(func() {
		mock.ExpectClose()
		Expect(db.Close()).To(Succeed())
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Describe("List", Label("list"), func() {
		var (
			now   time.Time
			rows  *sqlmock.Rows
			query string
		)

		BeforeEach(func() {
			now = time.Now().UTC().Truncate(time.Second)
			rows = sqlmock.NewRows([]string{"id", "text", "completed", "created_at", "updated_at"})
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
})
