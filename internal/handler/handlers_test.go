package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"person-api/internal/model"
	personsvc "person-api/internal/services/person"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPersonService struct{ mock.Mock }

func (m *MockPersonService) ListPersons(ctx context.Context, q model.PersonQuery) (model.PagedPersons, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(model.PagedPersons), args.Error(1)
}
func (m *MockPersonService) GetPersonByID(ctx context.Context, id int64) (model.Person, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Person), args.Error(1)
}
func (m *MockPersonService) CreatePerson(ctx context.Context, cmd model.CreatePersonCommand) (model.Person, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(model.Person), args.Error(1)
}
func (m *MockPersonService) UpdatePerson(ctx context.Context, id int64, cmd model.UpdatePersonCommand) (model.Person, error) {
	args := m.Called(ctx, id, cmd)
	return args.Get(0).(model.Person), args.Error(1)
}
func (m *MockPersonService) DeletePerson(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func setupRouter(s personsvc.Service) http.Handler {
	return NewRouter(s)
}

func TestHandleGetByID_Success(t *testing.T) {
	svc := new(MockPersonService)
	person := model.Person{ID: 1, Name: "John", Surname: "Doe"}
	svc.On("GetPersonByID", mock.Anything, int64(1)).Return(person, nil)

	req := httptest.NewRequest(http.MethodGet, "/persons/1", nil)
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var got model.Person
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.Equal(t, person, got)
	svc.AssertExpectations(t)
}

func TestHandleGetByID_InvalidID(t *testing.T) {
	svc := new(MockPersonService)

	req := httptest.NewRequest(http.MethodGet, "/persons/abc", nil)
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetByID_NotFound(t *testing.T) {
	svc := new(MockPersonService)
	svc.On("GetPersonByID", mock.Anything, int64(2)).Return(model.Person{}, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/persons/2", nil)
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestHandleCreate_Success(t *testing.T) {
	svc := new(MockPersonService)
	cmd := model.CreatePersonCommand{Name: "Jane", Surname: "Doe"}
	respPerson := model.Person{ID: 2, Name: "Jane", Surname: "Doe"}
	svc.On("CreatePerson", mock.Anything, cmd).Return(respPerson, nil)

	body, _ := json.Marshal(cmd)
	req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader(body))
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var got model.Person
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.Equal(t, respPerson, got)
	svc.AssertExpectations(t)
}

func TestHandleCreate_InvalidJSON(t *testing.T) {
	svc := new(MockPersonService)

	req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader([]byte("{invalid")))
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreate_ValidationError(t *testing.T) {
	svc := new(MockPersonService)
	req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewReader([]byte(`{"name":"","surname":""}`)))
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUpdate_NoFields(t *testing.T) {
	svc := new(MockPersonService)
	req := httptest.NewRequest(http.MethodPut, "/persons/1", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDelete_InvalidID(t *testing.T) {
	svc := new(MockPersonService)
	req := httptest.NewRequest(http.MethodDelete, "/persons/xyz", nil)
	w := httptest.NewRecorder()
	setupRouter(svc).ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParsePersonQuery_Default(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	q, err := parsePersonQuery(req)
	require.NoError(t, err)
	require.Equal(t, 1, q.Page)
	require.Equal(t, 10, q.PageSize)
}

func TestParsePersonQuery_AllFilters(t *testing.T) {
	params := url.Values{
		"page":        {"2"},
		"page_size":   {"5"},
		"name":        {"A"},
		"surname":     {"B"},
		"min_age":     {"10"},
		"max_age":     {"20"},
		"gender":      {"male"},
		"nationality": {"US"},
	}
	req := &http.Request{URL: &url.URL{RawQuery: params.Encode()}}
	q, err := parsePersonQuery(req)
	require.NoError(t, err)
	require.Equal(t, 2, q.Page)
	require.Equal(t, 5, q.PageSize)
	require.Equal(t, "A", *q.Name)
	require.Equal(t, "B", *q.Surname)
	require.Equal(t, 10, *q.MinAge)
	require.Equal(t, 20, *q.MaxAge)
	require.Equal(t, "male", *q.Gender)
	require.Equal(t, "US", *q.Nationality)
}

func TestParsePersonQuery_InvalidPage(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "page=0"}}
	_, err := parsePersonQuery(req)
	require.Error(t, err)
}

func TestParsePersonQuery_InvalidMinAge(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "min_age=-1"}}
	_, err := parsePersonQuery(req)
	require.Error(t, err)
}

func TestCreatePersonRequest_Validate(t *testing.T) {
	req := CreatePersonRequest{Name: "Ivan", Surname: "Petrov", Patronymic: ptr("Igorevich")}
	require.NoError(t, req.Validate())

	req2 := CreatePersonRequest{Name: "", Surname: "Petrov"}
	err := req2.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "name is required")

	req3 := CreatePersonRequest{Name: "123", Surname: "Petrov"}
	err = req3.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "name must contain only letters")
}

func TestUpdatePersonRequest_Validate(t *testing.T) {
	req := UpdatePersonRequest{Gender: ptr("female")}
	require.NoError(t, req.Validate())

	req2 := UpdatePersonRequest{Gender: ptr("unknown")}
	err := req2.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "gender must be 'male' or 'female'")

	neg := -5
	req3 := UpdatePersonRequest{Age: &neg}
	err = req3.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "age must be non-negative")

	code := "XYZ"
	req4 := UpdatePersonRequest{Nationality: &code}
	err = req4.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "nationality must be a 2-letter country code")
}

func ptr(text string) *string {
	return &text
}
