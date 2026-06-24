package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sumithsudesan/user-service/src/user"
)

// TestCreateUser tests the CreateUser handler of the user service.
func TestCreateUser(t *testing.T) {
	svc := user.NewService()
	handler := user.NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/users", handler.CreateUser)

	body := bytes.NewBufferString(`{"name":"testuser",
									"email":"test.user.2026@example.com",
									"status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, rr.Code)
	}
}
