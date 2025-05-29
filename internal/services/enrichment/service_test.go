package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"person-api/internal/model"

	"github.com/stretchr/testify/assert"
)

type stubTransport struct {
	responses map[string]*http.Response
	err       error
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if s.err != nil {
		return nil, s.err
	}
	if resp, ok := s.responses[req.URL.Host]; ok {
		return resp, nil
	}
	return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewReader(nil))}, nil
}

func makeResp(body interface{}, status int) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(http.Header),
	}
}

func TestEnrich_Success(t *testing.T) {
	ctx := context.Background()
	base := model.Person{Name: "Test"}
	st := &stubTransport{
		responses: map[string]*http.Response{
			"api.agify.io":     makeResp(map[string]int{"age": 25}, 200),
			"api.genderize.io": makeResp(map[string]string{"gender": "male"}, 200),
			"api.nationalize.io": makeResp(map[string][]map[string]interface{}{
				"country": {{"country_id": "GB", "probability": 0.5}},
			}, 200),
		},
	}
	svc := NewService().(*enrichmentService)
	svc.client = &http.Client{Transport: st}
	got, err := svc.Enrich(ctx, base)
	assert.NoError(t, err)
	assert.Equal(t, 25, *got.Age)
	assert.Equal(t, "male", *got.Gender)
	assert.Equal(t, "GB", *got.Nationality)
}

func TestEnrich_PartialNoCountry(t *testing.T) {
	ctx := context.Background()
	base := model.Person{Name: "X"}
	st := &stubTransport{
		responses: map[string]*http.Response{
			"api.agify.io":     makeResp(map[string]int{"age": 40}, 200),
			"api.genderize.io": makeResp(map[string]string{"gender": "female"}, 200),
			"api.nationalize.io": makeResp(map[string][]map[string]interface{}{
				"country": {},
			}, 200),
		},
	}
	svc := NewService().(*enrichmentService)
	svc.client = &http.Client{Transport: st}
	got, err := svc.Enrich(ctx, base)
	assert.NoError(t, err)
	assert.Nil(t, got.Nationality)
}

func TestEnrich_ErrorFromCall(t *testing.T) {
	ctx := context.Background()
	base := model.Person{Name: "Err"}
	st := &stubTransport{err: errors.New("network")}
	svc := NewService().(*enrichmentService)
	svc.client = &http.Client{Transport: st}
	_, err := svc.Enrich(ctx, base)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network")
}

func TestEnrich_Non200Status(t *testing.T) {
	ctx := context.Background()
	base := model.Person{Name: "Bad"}
	st := &stubTransport{
		responses: map[string]*http.Response{
			"api.agify.io": makeResp(map[string]int{"age": 50}, 500),
		},
	}
	svc := NewService().(*enrichmentService)
	svc.client = &http.Client{Transport: st}
	_, err := svc.Enrich(ctx, base)
	assert.Error(t, err)
}
