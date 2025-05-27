// internal/storage/postgres/postgres.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"person-api/internal/storage"
)

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) CreatePerson(ctx context.Context, p storage.PersonEntity) (storage.PersonEntity, error) {
	const q = `
    INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
    VALUES (:name, :surname, :patronymic, :age, :gender, :nationality)
    RETURNING id, created_at, updated_at`
	rows, err := s.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		return storage.PersonEntity{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return storage.PersonEntity{}, sql.ErrNoRows
	}
	if err := rows.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return storage.PersonEntity{}, err
	}
	return p, nil
}

func (s *PostgresStorage) UpdatePerson(ctx context.Context, id int64, p storage.PersonEntity) (storage.PersonEntity, error) {
	p.ID = id
	const q = `
    UPDATE persons SET
      name = :name,
      surname = :surname,
      patronymic = :patronymic,
      age = :age,
      gender = :gender,
      nationality = :nationality,
      updated_at = NOW()
    WHERE id = :id
    RETURNING created_at, updated_at`
	rows, err := s.db.NamedQueryContext(ctx, q, p)
	if err != nil {
		return storage.PersonEntity{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return storage.PersonEntity{}, sql.ErrNoRows
	}
	if err := rows.Scan(&p.CreatedAt, &p.UpdatedAt); err != nil {
		return storage.PersonEntity{}, err
	}
	return p, nil
}

func (s *PostgresStorage) DeletePerson(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM persons WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if cnt, _ := res.RowsAffected(); cnt == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresStorage) GetPersonByID(ctx context.Context, id int64) (storage.PersonEntity, error) {
	var p storage.PersonEntity
	const q = `
    SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
      FROM persons WHERE id=$1`
	if err := s.db.GetContext(ctx, &p, q, id); err != nil {
		return storage.PersonEntity{}, err
	}
	return p, nil
}

func (s *PostgresStorage) ListPersons(ctx context.Context, params storage.ListParams) (storage.PagedResult, error) {
	var conds []string
	var args []interface{}
	idx := 1
	if params.NameContains != nil {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", idx))
		args = append(args, "%"+*params.NameContains+"%")
		idx++
	}
	if params.SurnameContains != nil {
		conds = append(conds, fmt.Sprintf("surname ILIKE $%d", idx))
		args = append(args, "%"+*params.SurnameContains+"%")
		idx++
	}
	if params.MinAge != nil {
		conds = append(conds, fmt.Sprintf("age >= $%d", idx))
		args = append(args, *params.MinAge)
		idx++
	}
	if params.MaxAge != nil {
		conds = append(conds, fmt.Sprintf("age <= $%d", idx))
		args = append(args, *params.MaxAge)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM persons %s", where)
	if err := s.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return storage.PagedResult{}, err
	}

	dataQ := fmt.Sprintf(`
    SELECT id, name, surname, patronymic, age, gender, nationality, created_at, updated_at
      FROM persons %s ORDER BY id LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	args = append(args, params.Limit, params.Offset)

	var items []storage.PersonEntity
	if err := s.db.SelectContext(ctx, &items, dataQ, args...); err != nil {
		return storage.PagedResult{}, err
	}

	return storage.PagedResult{Items: items, TotalCount: total}, nil
}
