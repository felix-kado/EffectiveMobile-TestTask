package person

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/slog"

	"person-api/internal/model"
	"person-api/internal/services/enrichment"
	"person-api/internal/storage"
)

type mockEnr struct {
	mock.Mock
}

func (m *mockEnr) Enrich(ctx context.Context, p model.Person) (model.Person, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(model.Person), args.Error(1)
}

type mockStore struct {
	mock.Mock
}

func (m *mockStore) CreatePerson(ctx context.Context, p storage.PersonEntity) (storage.PersonEntity, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(storage.PersonEntity), args.Error(1)
}
func (m *mockStore) UpdatePerson(ctx context.Context, id int64, p storage.PersonEntity) (storage.PersonEntity, error) {
	args := m.Called(ctx, id, p)
	return args.Get(0).(storage.PersonEntity), args.Error(1)
}
func (m *mockStore) DeletePerson(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockStore) GetPersonByID(ctx context.Context, id int64) (storage.PersonEntity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(storage.PersonEntity), args.Error(1)
}
func (m *mockStore) ListPersons(ctx context.Context, params storage.ListParams) (storage.PagedResult, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(storage.PagedResult), args.Error(1)
}

func makeService(enr enrichment.Service, st storage.Storage) Service {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	return NewPersonService(logger, enr, st)
}

func TestCreatePerson_Success(t *testing.T) {
	ctx := context.Background()
	enrMock := new(mockEnr)
	storeMock := new(mockStore)

	cmd := model.CreatePersonCommand{Name: "John", Surname: "Doe", Patronymic: nil}
	enriched := model.Person{Name: "John", Surname: "Doe", Patronymic: nil, Age: intPtr(30), Gender: strPtr("male"), Nationality: strPtr("US")}
	enrMock.
		On("Enrich", ctx, model.Person{Name: cmd.Name, Surname: cmd.Surname, Patronymic: cmd.Patronymic}).
		Return(enriched, nil)
	inEntity := storage.PersonEntity{
		Name:        enriched.Name,
		Surname:     enriched.Surname,
		Patronymic:  enriched.Patronymic,
		Age:         enriched.Age,
		Gender:      enriched.Gender,
		Nationality: enriched.Nationality,
	}
	outEntity := inEntity
	outEntity.ID = 1
	storeMock.
		On("CreatePerson", ctx, inEntity).
		Return(outEntity, nil)

	svc := makeService(enrMock, storeMock)
	got, err := svc.CreatePerson(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
	assert.Equal(t, *enriched.Age, *got.Age)
	assert.Equal(t, *enriched.Gender, *got.Gender)
	assert.Equal(t, *enriched.Nationality, *got.Nationality)

	enrMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestCreatePerson_EnrichError(t *testing.T) {
	ctx := context.Background()
	enrMock := new(mockEnr)
	storeMock := new(mockStore)

	cmd := model.CreatePersonCommand{Name: "Jane", Surname: "Smith", Patronymic: nil}
	enrMock.
		On("Enrich", ctx, model.Person{Name: cmd.Name, Surname: cmd.Surname, Patronymic: cmd.Patronymic}).
		Return(model.Person{}, errors.New("api failure"))

	svc := makeService(enrMock, storeMock)
	_, err := svc.CreatePerson(ctx, cmd)

	assert.EqualError(t, err, "api failure")
	enrMock.AssertExpectations(t)
}

func TestUpdatePerson_Success(t *testing.T) {
	ctx := context.Background()
	enrMock := new(mockEnr)
	storeMock := new(mockStore)

	id := int64(42)
	old := storage.PersonEntity{ID: id, Name: "Old", Surname: "Name", Patronymic: nil, Age: intPtr(20), Gender: strPtr("female"), Nationality: strPtr("GB")}
	storeMock.On("GetPersonByID", ctx, id).Return(old, nil)

	cmd := model.UpdatePersonCommand{Name: strPtr("New"), Age: intPtr(25)}
	updatedEntity := old
	updatedEntity.Name = *cmd.Name
	updatedEntity.Age = cmd.Age

	outEntity := updatedEntity
	storeMock.On("UpdatePerson", ctx, id, updatedEntity).Return(outEntity, nil)

	svc := makeService(enrMock, storeMock)
	got, err := svc.UpdatePerson(ctx, id, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "New", got.Name)
	assert.Equal(t, 25, *got.Age)
	storeMock.AssertExpectations(t)
}

func TestUpdatePerson_GetError(t *testing.T) {
	ctx := context.Background()
	storeMock := new(mockStore)
	storeMock.On("GetPersonByID", ctx, int64(1)).Return(storage.PersonEntity{}, errors.New("not found"))

	svc := makeService(nil, storeMock)
	_, err := svc.UpdatePerson(ctx, 1, model.UpdatePersonCommand{})
	assert.EqualError(t, err, "not found")
}

func TestUpdatePerson_UpdateError(t *testing.T) {
	ctx := context.Background()
	storeMock := new(mockStore)
	id := int64(2)
	old := storage.PersonEntity{ID: id, Name: "A", Surname: "B"}
	storeMock.On("GetPersonByID", ctx, id).Return(old, nil)
	storeMock.On("UpdatePerson", ctx, id, mock.Anything).Return(storage.PersonEntity{}, errors.New("write error"))

	svc := makeService(nil, storeMock)
	_, err := svc.UpdatePerson(ctx, id, model.UpdatePersonCommand{Name: strPtr("X")})
	assert.EqualError(t, err, "write error")
}

func TestGetAndDeleteAndList(t *testing.T) {
	ctx := context.Background()
	storeMock := new(mockStore)

	entity := storage.PersonEntity{ID: 5, Name: "Foo", Surname: "Bar"}
	storeMock.On("GetPersonByID", ctx, int64(5)).Return(entity, nil)
	svc := makeService(nil, storeMock)
	p, err := svc.GetPersonByID(ctx, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), p.ID)

	storeMock.On("DeletePerson", ctx, int64(7)).Return(nil)
	err = svc.DeletePerson(ctx, 7)
	assert.NoError(t, err)

	params := storage.ListParams{Offset: 0, Limit: 10}
	paged := storage.PagedResult{
		Items:      []storage.PersonEntity{entity},
		TotalCount: 1,
	}
	storeMock.On("ListPersons", ctx, params).Return(paged, nil)
	res, err := svc.ListPersons(ctx, model.PersonQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Len(t, res.Persons, 1)

	storeMock.AssertExpectations(t)
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
