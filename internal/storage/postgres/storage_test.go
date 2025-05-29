// internal/storage/postgres/storage_test.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"person-api/internal/storage"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestCreatePerson_QuerySuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	// Expect INSERT with named params
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id, created_at, updated_at`)).
		WithArgs("A", "B", nil, nil, nil, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	ent := storage.PersonEntity{Name: "A", Surname: "B"}
	got, err := store.CreatePerson(context.Background(), ent)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreatePerson_NoRows(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	mock.ExpectQuery("INSERT INTO persons").
		WillReturnRows(sqlmock.NewRows([]string{}))
	_, err := store.CreatePerson(context.Background(), storage.PersonEntity{Name: "X"})
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestUpdatePerson_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}
	id := int64(2)
	created := time.Now()

	mock.ExpectQuery("UPDATE persons SET").
		WithArgs("A", "B", nil, nil, nil, nil, id).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(created, time.Now()))

	ent := storage.PersonEntity{Name: "A", Surname: "B"}
	got, err := store.UpdatePerson(context.Background(), id, ent)
	assert.NoError(t, err)
	assert.Equal(t, created.Format(time.RFC3339), got.CreatedAt.Format(time.RFC3339))
}

func TestUpdatePerson_NoRows(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	mock.ExpectQuery("UPDATE persons SET").
		WillReturnRows(sqlmock.NewRows([]string{}))
	_, err := store.UpdatePerson(context.Background(), 5, storage.PersonEntity{})
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeletePerson_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	mock.ExpectExec("DELETE FROM persons WHERE id=\\$1").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := store.DeletePerson(context.Background(), 3)
	assert.NoError(t, err)
}

func TestDeletePerson_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	mock.ExpectExec("DELETE FROM persons WHERE id=\\$1").
		WithArgs(4).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := store.DeletePerson(context.Background(), 4)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetPersonByID_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	cols := []string{"id", "name", "surname", "patronymic", "age", "gender", "nationality", "created_at", "updated_at"}
	mock.ExpectQuery("SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, "N", "S", nil, nil, nil, nil, time.Now(), time.Now()))

	ent, err := store.GetPersonByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), ent.ID)
}

func TestGetPersonByID_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	mock.ExpectQuery("SELECT id, name, surname").
		WillReturnError(fmt.Errorf("fail"))
	_, err := store.GetPersonByID(context.Background(), 2)
	assert.EqualError(t, err, "fail")
}

func TestListPersons_BaseAndFilter(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	store := &PostgresStorage{db: sqlxDB}

	// without filters
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM persons").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT id, name, surname").
		WithArgs(5, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender", "nationality", "created_at", "updated_at"}).
			AddRow(1, "A", "B", nil, nil, nil, nil, time.Now(), time.Now()).
			AddRow(2, "C", "D", nil, nil, nil, nil, time.Now(), time.Now()))
	// with a name filter
	params := storage.ListParams{
		NameContains: ptrString("A"),
		Offset:       0,
		Limit:        5,
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM persons WHERE name ILIKE $1")).
		WithArgs("%A%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at FROM persons WHERE name ILIKE $1 ORDER BY id LIMIT $2 OFFSET $3")).
		WithArgs("%A%", 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "surname", "patronymic", "age", "gender", "nationality", "created_at", "updated_at"}).
			AddRow(1, "A", "B", nil, nil, nil, nil, time.Now(), time.Now()))

	// run no-filter
	_, err := store.ListPersons(context.Background(), storage.ListParams{Offset: 0, Limit: 5})
	assert.NoError(t, err)
	// run with filter
	_, err = store.ListPersons(context.Background(), params)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func ptrString(s string) *string { return &s }
