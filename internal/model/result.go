package model

type Person struct {
	ID          int64
	Name        string
	Surname     string
	Patronymic  *string
	Age         *int
	Gender      *string
	Nationality *string
	CreatedAt   string
}

type PagedPersons struct {
	Persons  []Person
	Total    int
	Page     int
	PageSize int
}
