package model

type CreatePersonCommand struct {
	Name       string
	Surname    string
	Patronymic *string
}

type UpdatePersonCommand struct {
	Name       *string
	Surname    *string
	Patronymic *string
}
