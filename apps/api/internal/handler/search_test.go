package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorolens/sorolens/apps/api/internal/store"
)

func TestSearch(t *testing.T) {
	t.Run("returns matching results", func(t *testing.T) {
		ms := store.NewMockStore()
		if err := ms.UpsertContract(t.Context(), store.Contract{
			ID:      "C12345",
			Label:   "Token contract",
			Network: "testnet",
		}); err != nil {
			t.Fatalf("UpsertContract: %v", err)
		}

		srv := newTestHandler(ms, true, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=token", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status: got %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var response struct {
			Items []struct {
				Type    string `json:"type"`
				ID      string `json:"id"`
				Label   string `json:"label"`
				Network string `json:"network"`
			} `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Items) != 1 {
			t.Fatalf("got %d results, want 1", len(response.Items))
		}
		got := response.Items[0]
		if got.Type != "contract" || got.ID != "C12345" || got.Label != "Token contract" || got.Network != "testnet" {
			t.Errorf("unexpected result: %+v", got)
		}
	})

	t.Run("rejects an empty query", func(t *testing.T) {
		srv := newTestHandler(store.NewMockStore(), true, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=%20%20", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status: got %d, want %d; body: %s", w.Code, http.StatusUnprocessableEntity, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "q is required") {
			t.Errorf("response body %q does not contain the validation message", w.Body.String())
		}
	})

	t.Run("returns an error when the store fails", func(t *testing.T) {
		ms := store.NewMockStore()
		ms.SearchErr = errors.New("search unavailable")
		srv := newTestHandler(ms, true, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=token", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status: got %d, want %d; body: %s", w.Code, http.StatusInternalServerError, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "failed to search") {
			t.Errorf("response body %q does not contain the error message", w.Body.String())
		}
	})
}
