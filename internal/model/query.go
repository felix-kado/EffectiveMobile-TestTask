package model

type PersonQuery struct {
	Name        *string
	Surname     *string
	Gender      *string
	Nationality *string
	MinAge      *int
	MaxAge      *int
	Page        int
	PageSize    int
}
