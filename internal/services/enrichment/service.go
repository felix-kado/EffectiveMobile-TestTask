package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"person-api/internal/model"
)

type Service interface {
	Enrich(ctx context.Context, p model.Person) (model.Person, error)
}

type enrichmentService struct {
	client *http.Client
}

func NewService() Service {
	return &enrichmentService{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *enrichmentService) Enrich(ctx context.Context, p model.Person) (model.Person, error) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
	)

	wg.Add(3)
	// agify
	go func() {
		defer wg.Done()
		var a struct {
			Age int `json:"age"`
		}
		if err := s.call(ctx, fmt.Sprintf("https://api.agify.io/?name=%s", p.Name), &a); err != nil {
			mu.Lock()
			if first == nil {
				first = err
			}
			mu.Unlock()
			return
		}
		mu.Lock()
		p.Age = &a.Age
		mu.Unlock()
	}()

	// genderize
	go func() {
		defer wg.Done()
		var g struct {
			Gender string `json:"gender"`
		}
		if err := s.call(ctx, fmt.Sprintf("https://api.genderize.io/?name=%s", p.Name), &g); err != nil {
			mu.Lock()
			if first == nil {
				first = err
			}
			mu.Unlock()
			return
		}
		mu.Lock()
		p.Gender = &g.Gender
		mu.Unlock()
	}()

	// nationalize
	go func() {
		defer wg.Done()
		var n struct {
			Country []struct {
				CountryID   string  `json:"country_id"`
				Probability float64 `json:"probability"`
			} `json:"country"`
		}
		if err := s.call(ctx, fmt.Sprintf("https://api.nationalize.io/?name=%s", p.Name), &n); err != nil {
			mu.Lock()
			if first == nil {
				first = err
			}
			mu.Unlock()
			return
		}
		if len(n.Country) > 0 {
			mu.Lock()
			p.Nationality = &n.Country[0].CountryID
			mu.Unlock()
		}
	}()

	wg.Wait()
	if first != nil {
		return p, first
	}
	return p, nil
}

func (s *enrichmentService) call(ctx context.Context, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
