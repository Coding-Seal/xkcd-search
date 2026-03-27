package templates

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplates_Pics(t *testing.T) {
	w := httptest.NewRecorder()
	err := Pics(w, PicsParams{
		Comics: []Comic{{ID: 1, Title: "Test Comic", ImgURL: "http://example.com/1.png"}},
		Layout: Layout{PageTitle: "Search"},
	})
	assert.NoError(t, err)
}

func TestTemplates_Pics_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	err := Pics(w, PicsParams{Comics: []Comic{}})
	assert.NoError(t, err)
}

func TestTemplates_Login(t *testing.T) {
	w := httptest.NewRecorder()
	err := Login(w, LoginParams{Layout: Layout{PageTitle: "Login"}})
	assert.NoError(t, err)
}

func TestTemplates_Update(t *testing.T) {
	w := httptest.NewRecorder()
	err := Update(w, UpdateParams{Layout: Layout{PageTitle: "Update"}})
	assert.NoError(t, err)
}

func TestTemplates_Favorites(t *testing.T) {
	w := httptest.NewRecorder()
	err := Favorites(w, FavoritesParams{
		Comics: []Comic{
			{ID: 2, Title: "Fav Comic", ImgURL: "http://example.com/2.png", Favorite: true},
		},
		Layout: Layout{PageTitle: "Favorites"},
	})
	assert.NoError(t, err)
}

func TestTemplates_Favorites_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	err := Favorites(w, FavoritesParams{Comics: []Comic{}})
	assert.NoError(t, err)
}

func TestTemplates_Main(t *testing.T) {
	w := httptest.NewRecorder()
	err := Main(w, MainParams{Layout: Layout{PageTitle: "Home"}})
	assert.NoError(t, err)
}
