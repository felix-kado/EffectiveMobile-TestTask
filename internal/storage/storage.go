package storage

import (
	"context"
	"time"
)

type PersonEntity struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Surname     string    `db:"surname"`
	Patronymic  *string   `db:"patronymic"`
	Age         *int      `db:"age"`
	Gender      *string   `db:"gender"`
	Nationality *string   `db:"nationality"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ListParams struct {
	NameContains    *string
	SurnameContains *string
	MinAge          *int
	MaxAge          *int
	Offset          int
	Limit           int
}

type PagedResult struct {
	Items      []PersonEntity
	TotalCount int64
}

type Storage interface {
	CreatePerson(ctx context.Context, p PersonEntity) (PersonEntity, error)
	UpdatePerson(ctx context.Context, id int64, p PersonEntity) (PersonEntity, error)
	DeletePerson(ctx context.Context, id int64) error
	GetPersonByID(ctx context.Context, id int64) (PersonEntity, error)
	ListPersons(ctx context.Context, params ListParams) (PagedResult, error)
}
