package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"person-api/internal/model"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// логируем прямо в консоль, т.к. логгер тут недоступен
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parsePersonQuery(r *http.Request) (model.PersonQuery, error) {
	q := model.PersonQuery{Page: 1, PageSize: 10}
	var err error

	if v := r.URL.Query().Get("page"); v != "" {
		q.Page, err = strconv.Atoi(v)
		if err != nil || q.Page < 1 {
			return q, errors.New("invalid page parameter")
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		q.PageSize, err = strconv.Atoi(v)
		if err != nil || q.PageSize < 1 {
			return q, errors.New("invalid page_size parameter")
		}
	}
	// остальные фильтры без ошибок: строки и неотрицательные числа
	if v := r.URL.Query().Get("name"); v != "" {
		q.Name = &v
	}
	if v := r.URL.Query().Get("surname"); v != "" {
		q.Surname = &v
	}
	if v := r.URL.Query().Get("min_age"); v != "" {
		x, err2 := strconv.Atoi(v)
		if err2 != nil || x < 0 {
			return q, errors.New("invalid min_age parameter")
		}
		q.MinAge = &x
	}
	if v := r.URL.Query().Get("max_age"); v != "" {
		x, err2 := strconv.Atoi(v)
		if err2 != nil || x < 0 {
			return q, errors.New("invalid max_age parameter")
		}
		q.MaxAge = &x
	}
	if v := r.URL.Query().Get("gender"); v != "" {
		q.Gender = &v
	}
	if v := r.URL.Query().Get("nationality"); v != "" {
		q.Nationality = &v
	}
	return q, nil
}
