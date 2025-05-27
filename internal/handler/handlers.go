package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"person-api/internal/model"
	"strconv"

	"github.com/go-chi/chi/v5"
	"person-api/internal/services/person"
)

// @Summary      List persons
// @Description  Returns paginated list of persons with optional filters
// @Tags         persons
// @Accept       json
// @Produce      json
// @Param        page         query   int              false  "Page number"       default(1)
// @Param        page_size    query   int              false  "Items per page"    default(10)
// @Param        name         query   string           false  "Filter by name"
// @Param        surname      query   string           false  "Filter by surname"
// @Param        min_age      query   int              false  "Filter by minimum age"
// @Param        max_age      query   int              false  "Filter by maximum age"
// @Param        gender       query   string           false  "Filter by gender"
// @Param        nationality  query   string           false  "Filter by nationality"
// @Success      200  {object}  PagedPersonsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /persons [get]
func handleList(svc person.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := parsePersonQuery(r)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		res, err := svc.ListPersons(r.Context(), q)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "cannot list persons")
			return
		}

		out := PagedPersonsResponse{
			Persons:  make([]PersonResponse, len(res.Persons)),
			Total:    res.Total,
			Page:     res.Page,
			PageSize: res.PageSize,
		}
		for i, p := range res.Persons {
			out.Persons[i] = PersonResponse(p)
		}
		respondJSON(w, http.StatusOK, out)
	}
}

// @Summary      Get person by ID
// @Description  Returns a single person by their ID
// @Tags         persons
// @Accept       json
// @Produce      json
// @Param        id   path      int              true   "Person ID"
// @Success      200  {object}  PersonResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /persons/{id} [get]
func handleGetByID(svc person.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id < 1 {
			respondError(w, http.StatusBadRequest, "invalid id")
			return
		}

		p, err := svc.GetPersonByID(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "person not found")
			return
		}
		respondJSON(w, http.StatusOK, PersonResponse(p))
	}
}

// @Summary      Create person
// @Description  Creates a new person and enriches their data (age, gender, nationality)
// @Tags         persons
// @Accept       json
// @Produce      json
// @Param        payload  body      CreatePersonRequest  true   "Person payload"
// @Success      201      {object}  PersonResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /persons [post]
func handleCreate(svc person.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreatePersonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request payload")
			return
		}
		// простая валидация полей
		if req.Name == "" || req.Surname == "" {
			respondError(w, http.StatusBadRequest, "name and surname are required")
			return
		}

		cmd := model.CreatePersonCommand{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}
		p, err := svc.CreatePerson(r.Context(), cmd)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "could not create person")
			return
		}
		respondJSON(w, http.StatusCreated, PersonResponse(p))
	}
}

// @Summary      Update person
// @Description  Updates one or more fields of an existing person
// @Tags         persons
// @Accept       json
// @Produce      json
// @Param        id       path      int                  true   "Person ID"
// @Param        payload  body      UpdatePersonRequest  true   "Fields to update"
// @Success      200      {object}  PersonResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /persons/{id} [put]
func handleUpdate(svc person.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id < 1 {
			respondError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req UpdatePersonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request payload")
			return
		}
		if req.Name == nil && req.Surname == nil && req.Patronymic == nil {
			respondError(w, http.StatusBadRequest, "no fields to update")
			return
		}

		cmd := model.UpdatePersonCommand{
			Name:       req.Name,
			Surname:    req.Surname,
			Patronymic: req.Patronymic,
		}
		p, err := svc.UpdatePerson(r.Context(), id, cmd)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondError(w, http.StatusNotFound, "person not found")
			} else {
				respondError(w, http.StatusInternalServerError, "could not update person")
			}
			return
		}
		respondJSON(w, http.StatusOK, PersonResponse(p))
	}
}

// @Summary      Delete person
// @Description  Deletes a person by their ID
// @Tags         persons
// @Accept       json
// @Produce      json
// @Param        id   path      int              true   "Person ID"
// @Success      204  {string}  string            "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /persons/{id} [delete]
func handleDelete(svc person.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id < 1 {
			respondError(w, http.StatusBadRequest, "invalid id")
			return
		}
		if err = svc.DeletePerson(r.Context(), id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondError(w, http.StatusNotFound, "person not found")
			} else {
				respondError(w, http.StatusInternalServerError, "could not delete person")
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
