//nolint:gochecknoinits
package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"yadro-go-course/web/rest"
)

func TestNewClient(t *testing.T) {
	c := rest.NewClient("http://localhost:8080")
	assert.NotNil(t, c)
}

func TestClient_Login_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Authorization", "test-jwt-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	token, err := c.Login(context.Background(), "admin", "secret")
	require.NoError(t, err)
	assert.Equal(t, "test-jwt-token", token)
}

func TestClient_Login_WrongCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	token, err := c.Login(context.Background(), "admin", "wrong")
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestClient_Login_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	_, err := c.Login(context.Background(), "user", "pass")
	assert.Error(t, err)
}

func TestClient_Login_NetworkError(t *testing.T) {
	c := rest.NewClient("http://127.0.0.1:1") // nothing listens here
	_, err := c.Login(context.Background(), "user", "pass")
	assert.Error(t, err)
}

func TestClient_SearchPics_Found(t *testing.T) {
	comics := []rest.Comic{
		{ID: 1, Title: "Test Comic", ImgURL: "http://example.com/1.png"},
		{ID: 2, Title: "Another Comic", ImgURL: "http://example.com/2.png"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "funny", r.URL.Query().Get("search"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comics)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	result, err := c.SearchPics(context.Background(), "funny")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
}

func TestClient_SearchPics_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	result, err := c.SearchPics(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestClient_SearchPics_ServerError(t *testing.T) {
	c := rest.NewClient("http://127.0.0.1:1")
	_, err := c.SearchPics(context.Background(), "test")
	assert.Error(t, err)
}

func TestClient_SearchPics_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid json"))
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	_, err := c.SearchPics(context.Background(), "test")
	assert.Error(t, err)
}

func TestClient_Update_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "my-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	err := c.Update(context.Background(), "my-token")
	assert.NoError(t, err)
}

func TestClient_Update_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	err := c.Update(context.Background(), "bad-token")
	assert.Error(t, err)
}

func TestClient_Update_NetworkError(t *testing.T) {
	c := rest.NewClient("http://127.0.0.1:1")
	err := c.Update(context.Background(), "token")
	assert.Error(t, err)
}

func TestClient_Comic_Found(t *testing.T) {
	expected := rest.Comic{ID: 42, Title: "Deep Time", ImgURL: "http://example.com/42.png"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	got, err := c.Comic(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, got.ID)
	assert.Equal(t, expected.Title, got.Title)
}

func TestClient_Comic_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := rest.NewClient(srv.URL)
	_, err := c.Comic(context.Background(), 999)
	assert.Error(t, err)
}

func TestClient_Comic_NetworkError(t *testing.T) {
	c := rest.NewClient("http://127.0.0.1:1")
	_, err := c.Comic(context.Background(), 1)
	assert.Error(t, err)
}
