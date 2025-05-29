package person

import (
	"context"
	"person-api/internal/services/enrichment"
	"time"

	"golang.org/x/exp/slog"
	"person-api/internal/model"
	"person-api/internal/storage"
)

type Service interface {
	CreatePerson(ctx context.Context, cmd model.CreatePersonCommand) (model.Person, error)
	UpdatePerson(ctx context.Context, id int64, cmd model.UpdatePersonCommand) (model.Person, error)
	DeletePerson(ctx context.Context, id int64) error
	GetPersonByID(ctx context.Context, id int64) (model.Person, error)
	ListPersons(ctx context.Context, q model.PersonQuery) (model.PagedPersons, error)
}

type personService struct {
	logger slog.Logger
	es     enrichment.Service
	st     storage.Storage
}

func NewPersonService(logger *slog.Logger, es enrichment.Service, st storage.Storage) Service {
	return &personService{logger: *logger, es: es, st: st}
}

func (s *personService) CreatePerson(ctx context.Context, cmd model.CreatePersonCommand) (model.Person, error) {
	s.logger.Info("CreatePerson", "cmd", cmd)
	pr := model.Person{Name: cmd.Name, Surname: cmd.Surname, Patronymic: cmd.Patronymic}
	enriched, err := s.es.Enrich(ctx, pr)
	if err != nil {
		return model.Person{}, err
	}
	pe := storage.PersonEntity{
		Name:        enriched.Name,
		Surname:     enriched.Surname,
		Patronymic:  enriched.Patronymic,
		Age:         enriched.Age,
		Gender:      enriched.Gender,
		Nationality: enriched.Nationality,
	}
	saved, err := s.st.CreatePerson(ctx, pe)
	if err != nil {
		return model.Person{}, err
	}
	return mapEntity(saved), nil
}

func (s *personService) UpdatePerson(ctx context.Context, id int64, cmd model.UpdatePersonCommand) (model.Person, error) {
	s.logger.Info("UpdatePerson", "id", id, "cmd", cmd)
	old, err := s.st.GetPersonByID(ctx, id)
	if err != nil {
		return model.Person{}, err
	}
	if cmd.Name != nil {
		old.Name = *cmd.Name
	}
	if cmd.Surname != nil {
		old.Surname = *cmd.Surname
	}
	if cmd.Patronymic != nil {
		old.Patronymic = cmd.Patronymic
	}
	if cmd.Age != nil {
		old.Age = cmd.Age
	}
	if cmd.Gender != nil {
		old.Gender = cmd.Gender
	}
	if cmd.Nationality != nil {
		old.Nationality = cmd.Nationality
	}
	updated, err := s.st.UpdatePerson(ctx, id, old)
	if err != nil {
		return model.Person{}, err
	}
	return mapEntity(updated), nil
}

func (s *personService) DeletePerson(ctx context.Context, id int64) error {
	s.logger.Info("DeletePerson", "id", id)
	return s.st.DeletePerson(ctx, id)
}

func (s *personService) GetPersonByID(ctx context.Context, id int64) (model.Person, error) {
	s.logger.Info("GetPersonByID", "id", id)
	e, err := s.st.GetPersonByID(ctx, id)
	if err != nil {
		return model.Person{}, err
	}
	return mapEntity(e), nil
}

func (s *personService) ListPersons(ctx context.Context, q model.PersonQuery) (model.PagedPersons, error) {
	s.logger.Info("ListPersons", "query", q)
	params := storage.ListParams{
		Offset: q.PageSize * (q.Page - 1),
		Limit:  q.PageSize,
	}
	params.NameContains = q.Name
	params.SurnameContains = q.Surname
	params.MinAge = q.MinAge
	params.MaxAge = q.MaxAge
	res, err := s.st.ListPersons(ctx, params)
	if err != nil {
		return model.PagedPersons{}, err
	}
	out := make([]model.Person, len(res.Items))
	for i, e := range res.Items {
		out[i] = mapEntity(e)
	}
	return model.PagedPersons{
		Persons:  out,
		Total:    int(res.TotalCount),
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

func mapEntity(e storage.PersonEntity) model.Person {
	return model.Person{
		ID:          e.ID,
		Name:        e.Name,
		Surname:     e.Surname,
		Patronymic:  e.Patronymic,
		Age:         e.Age,
		Gender:      e.Gender,
		Nationality: e.Nationality,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
	}
}
