package handler

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// разрешаем только буквы русского и латинского алфавита
	letterRegex = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё]+$`)
	// для nationality — двухбуквенный код страны в верхнем регистре
	nationalityRegex = regexp.MustCompile(`^[A-Z]{2}$`)
)

// Validate implements validation for CreatePersonRequest.
func (r CreatePersonRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name,
			validation.Required.Error("name is required"),
			validation.Match(letterRegex).Error("name must contain only letters"),
		),
		validation.Field(&r.Surname,
			validation.Required.Error("surname is required"),
			validation.Match(letterRegex).Error("surname must contain only letters"),
		),
		validation.Field(&r.Patronymic,
			validation.When(r.Patronymic != nil && *r.Patronymic != "", validation.Match(letterRegex).Error("patronymic must contain only letters")),
		),
	)
}

// Validate implements validation for UpdatePersonRequest.
func (r UpdatePersonRequest) Validate() error {
	return validation.ValidateStruct(&r,
		// имя
		validation.Field(&r.Name,
			validation.When(r.Name != nil,
				validation.Match(letterRegex).Error("name must contain only letters"),
			),
		),
		// фамилия
		validation.Field(&r.Surname,
			validation.When(r.Surname != nil,
				validation.Match(letterRegex).Error("surname must contain only letters"),
			),
		),
		// отчество
		validation.Field(&r.Patronymic,
			validation.When(r.Patronymic != nil,
				validation.Match(letterRegex).Error("patronymic must contain only letters"),
			),
		),
		// возраст
		validation.Field(&r.Age,
			validation.When(r.Age != nil,
				validation.Min(0).Error("age must be non-negative"),
			),
		),
		// пол
		validation.Field(&r.Gender,
			validation.When(r.Gender != nil,
				validation.In("male", "female").Error("gender must be 'male' or 'female'"),
			),
		),
		// национальность
		validation.Field(&r.Nationality,
			validation.When(r.Nationality != nil,
				validation.Match(nationalityRegex).Error("nationality must be a 2-letter country code"),
			),
		),
	)
}
