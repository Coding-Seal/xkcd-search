//nolint:gochecknoinits
package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"yadro-go-course/web/handlers"
	"yadro-go-course/web/rest"
)

// failWriter simulates a ResponseWriter whose Write method always fails.
type failWriter struct{ hdr http.Header }

func newFailWriter() *failWriter                { return &failWriter{hdr: http.Header{}} }
func (f *failWriter) Header() http.Header       { return f.hdr }
func (f *failWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("forced write error") }
func (f *failWriter) WriteHeader(int)           {}

// apiServer creates a mock API server with configurable handlers per path.
func apiServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestMainHandler_Success(t *testing.T) {
	h := handlers.MainHandler()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPics_NoResults(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rest.NewClient(srv.URL)
	h := handlers.Pics(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?query=nothing", nil)
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPics_WithComics(t *testing.T) {
	comics := []map[string]interface{}{
		{"id": 1, "title": "Test Comic", "img": "http://example.com/1.png"},
	}

	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comics)
	}))
	defer srv.Close()

	client := rest.NewClient(srv.URL)
	h := handlers.Pics(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?query=test", nil)
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPics_WithFavoritesCookie(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	favJSON, _ := json.Marshal([]int{1, 2})
	client := rest.NewClient(srv.URL)
	h := handlers.Pics(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?query=test", nil)
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: string(favJSON)})
	err := h(w, r)
	assert.NoError(t, err)
}

func TestPics_InvalidFavoritesCookie(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := rest.NewClient(srv.URL)
	h := handlers.Pics(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?query=test", nil)
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: "not-json"})
	err := h(w, r)
	// Invalid cookie → bad request error
	assert.Error(t, err)
}

func TestLogin_GET(t *testing.T) {
	h := handlers.Login()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLoginForm_EmptyCredentials(t *testing.T) {
	client := rest.NewClient("http://127.0.0.1:1") // won't be called
	h := handlers.LoginForm(client)
	w := httptest.NewRecorder()
	form := url.Values{"login": {""}, "pswd": {""}}
	r := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := h(w, r)
	// Should redirect with error – not nil because ErrNoLoginOrPassword is returned
	assert.Error(t, err)
}

func TestLoginForm_Success(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "valid-jwt-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := rest.NewClient(srv.URL)
	h := handlers.LoginForm(client)
	w := httptest.NewRecorder()
	form := url.Values{"login": {"admin"}, "pswd": {"secret"}}
	r := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := h(w, r)
	assert.NoError(t, err)
}

func TestFavorites_Empty(t *testing.T) {
	client := rest.NewClient("http://127.0.0.1:1")
	h := handlers.Favorites(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/favorites", nil)
	// No Favorites cookie → empty list → renders template
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestToggleFavorites_AddNew(t *testing.T) {
	h := handlers.ToggleFavorites()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/favorites/toggle/1", nil)
	r.SetPathValue("id", "1")
	err := h(w, r)
	assert.NoError(t, err)
	// Verify a Favorites cookie was set
	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies)
	var found bool
	for _, c := range cookies {
		if c.Name == "Favorites" {
			found = true
			var ids []int
			json.Unmarshal([]byte(c.Value), &ids)
			assert.Contains(t, ids, 1)
		}
	}
	assert.True(t, found, "Favorites cookie should be set")
}

func TestToggleFavorites_Remove(t *testing.T) {
	h := handlers.ToggleFavorites()
	w := httptest.NewRecorder()
	favJSON, _ := json.Marshal([]int{1, 2, 3})
	r := httptest.NewRequest("POST", "/favorites/toggle/2", nil)
	r.SetPathValue("id", "2")
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: string(favJSON)})
	err := h(w, r)
	assert.NoError(t, err)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "Favorites" {
			var ids []int
			json.Unmarshal([]byte(c.Value), &ids)
			assert.NotContains(t, ids, 2)
		}
	}
}

func TestToggleFavorites_InvalidID(t *testing.T) {
	h := handlers.ToggleFavorites()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/favorites/toggle/abc", nil)
	r.SetPathValue("id", "abc")
	err := h(w, r)
	assert.Error(t, err)
}

func TestFavorites_WithComics(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "title": "Test", "img": "http://example.com/1.png"})
	}))
	defer srv.Close()

	favJSON, _ := json.Marshal([]int{1})
	client := rest.NewClient(srv.URL)
	h := handlers.Favorites(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/favorites", nil)
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: string(favJSON)})
	err := h(w, r)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFavorites_ComicError(t *testing.T) {
	srv := apiServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	favJSON, _ := json.Marshal([]int{1})
	client := rest.NewClient(srv.URL)
	h := handlers.Favorites(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/favorites", nil)
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: string(favJSON)})
	err := h(w, r)
	assert.Error(t, err)
}

func TestFavorites_InvalidCookie(t *testing.T) {
	client := rest.NewClient("http://127.0.0.1:1")
	h := handlers.Favorites(client)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/favorites", nil)
	r.AddCookie(&http.Cookie{Name: "Favorites", Value: "not-valid-json"})
	err := h(w, r)
	assert.Error(t, err)
}

func TestMainHandler_TemplateError(t *testing.T) {
	h := handlers.MainHandler()
	r := httptest.NewRequest("GET", "/", nil)
	err := h(newFailWriter(), r)
	assert.Error(t, err)
}

func TestLogin_TemplateError(t *testing.T) {
	h := handlers.Login()
	r := httptest.NewRequest("GET", "/login", nil)
	err := h(newFailWriter(), r)
	assert.Error(t, err)
}
