package model

type CreatePersonCommand struct {
	Name       string
	Surname    string
	Patronymic *string
}

type UpdatePersonCommand struct {
	Name        *string
	Surname     *string
	Patronymic  *string
	Age         *int
	Gender      *string
	Nationality *string
}
