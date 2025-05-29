package handler

type ErrorResponse struct {
	Error string `json:"error"`
}
type CreatePersonRequest struct {
	Name       string  `json:"name" validate:"required"`
	Surname    string  `json:"surname" validate:"required"`
	Patronymic *string `json:"patronymic"`
}

type UpdatePersonRequest struct {
	Name        *string `json:"name"`
	Surname     *string `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	Age         *int    `json:"age"`
	Gender      *string `json:"gender"`
	Nationality *string `json:"nationality"`
}

type PersonResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	Patronymic  *string `json:"patronymic"`
	Age         *int    `json:"age"`
	Gender      *string `json:"gender"`
	Nationality *string `json:"nationality"`
	CreatedAt   string  `json:"created_at"`
}

type PagedPersonsResponse struct {
	Persons  []PersonResponse `json:"persons"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
