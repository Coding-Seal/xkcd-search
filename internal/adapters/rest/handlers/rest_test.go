package handlers

import (
	"net/http/httptest"
	"testing"

	http_util "yadro-go-course/pkg/http-util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"yadro-go-course/internal/adapters/repos/fetcher"
	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
	"yadro-go-course/internal/core/services"
	test "yadro-go-course/test/fetcher"
)

func TestUpdate_Happy(t *testing.T) {
	numTestComic := 10

	var comicRepo comicRepoMock

	var searchRepo searchRepoMock

	for i := 1; i <= numTestComic/2; i++ {
		comicRepo.On("Comic", i).Return(models.Comic{ID: i}, nil)
	}
	comicRepo.On("Comic", mock.Anything).Return(models.Comic{}, ports.ErrNotFound)
	comicRepo.On("Store", mock.Anything).Return(nil)
	searchRepo.On("AddComic", mock.Anything)

	srv := test.NewMockXKCD(numTestComic)
	fetcherRepo := fetcher.NewFetcher(srv.URL, numTestComic)
	t.Cleanup(srv.Close)

	fetcherSrv := services.NewFetcher(fetcherRepo, &comicRepo, &searchRepo)
	h := Update(fetcherSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := h(w, r)
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	numTestComic := 10

	var comicRepo comicRepoMock

	var searchRepo searchRepoMock

	for i := 1; i <= numTestComic/2; i++ {
		comicRepo.On("Comic", i).Return(models.Comic{ID: i}, nil)
	}
	comicRepo.On("Comic", mock.Anything).Return(models.Comic{}, ports.ErrNotFound)
	comicRepo.On("Store", mock.Anything).Return(ports.ErrInternal)
	searchRepo.On("AddComic", mock.Anything)

	srv := test.NewMockXKCD(numTestComic)
	fetcherRepo := fetcher.NewFetcher(srv.URL, numTestComic)
	t.Cleanup(srv.Close)

	fetcherSrv := services.NewFetcher(fetcherRepo, &comicRepo, &searchRepo)
	h := Update(fetcherSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrInternal)
}

func TestSearch_Found(t *testing.T) {
	var comicRepo comicRepoMock
	var searchRepo searchRepoMock

	searchRepo.On("SearchComics", "funny").Return(map[int]int{1: 5, 2: 3})
	comicRepo.On("Comic", 1).Return(models.Comic{ID: 1, Title: "Funny Comic", ImgURL: "http://imgs.xkcd.com/1.png"}, nil)
	comicRepo.On("Comic", 2).Return(models.Comic{ID: 2, Title: "Another One", ImgURL: "http://imgs.xkcd.com/2.png"}, nil)

	searchSrv := services.NewSearch(&searchRepo, &comicRepo)
	h := Search(searchSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?search=funny", nil)
	err := h(w, r)
	assert.NoError(t, err)
}

func TestSearch_NotFound(t *testing.T) {
	var comicRepo comicRepoMock
	var searchRepo searchRepoMock

	searchRepo.On("SearchComics", "nonexistent").Return(map[int]int{})

	searchSrv := services.NewSearch(&searchRepo, &comicRepo)
	h := Search(searchSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?search=nonexistent", nil)
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrNotFound)
}

func TestSearch_EmptyQuery(t *testing.T) {
	var comicRepo comicRepoMock
	var searchRepo searchRepoMock

	searchRepo.On("SearchComics", "").Return(map[int]int{})

	searchSrv := services.NewSearch(&searchRepo, &comicRepo)
	h := Search(searchSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrNotFound)
}

func TestSearch_SingleResult(t *testing.T) {
	var comicRepo comicRepoMock
	var searchRepo searchRepoMock

	searchRepo.On("SearchComics", "science").Return(map[int]int{42: 10})
	comicRepo.On("Comic", 42).Return(models.Comic{ID: 42, Title: "Science!", ImgURL: "http://imgs.xkcd.com/42.png"}, nil)

	searchSrv := services.NewSearch(&searchRepo, &comicRepo)
	h := Search(searchSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?search=science", nil)
	err := h(w, r)
	assert.NoError(t, err)
}

func TestComic_Found(t *testing.T) {
	var comicRepo comicRepoMock

	comicRepo.On("Comic", 1).Return(models.Comic{ID: 1, Title: "Test Comic", ImgURL: "http://imgs.xkcd.com/1.png"}, nil)

	comicSrv := services.NewComicService(&comicRepo)
	h := Comic(comicSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/comic/1", nil)
	r.SetPathValue("id", "1")
	err := h(w, r)
	assert.NoError(t, err)
}

func TestComic_InvalidID(t *testing.T) {
	var comicRepo comicRepoMock

	comicSrv := services.NewComicService(&comicRepo)
	h := Comic(comicSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/comic/abc", nil)
	r.SetPathValue("id", "abc")
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrBadRequest)
}

func TestComic_ServiceError(t *testing.T) {
	var comicRepo comicRepoMock

	comicRepo.On("Comic", 999).Return(models.Comic{}, ports.ErrNotFound)

	comicSrv := services.NewComicService(&comicRepo)
	h := Comic(comicSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/comic/999", nil)
	r.SetPathValue("id", "999")
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrInternal)
}

func TestComic_ZeroID(t *testing.T) {
	var comicRepo comicRepoMock

	comicRepo.On("Comic", 0).Return(models.Comic{}, ports.ErrNotFound)

	comicSrv := services.NewComicService(&comicRepo)
	h := Comic(comicSrv)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/comic/0", nil)
	r.SetPathValue("id", "0")
	err := h(w, r)
	assert.ErrorIs(t, err, http_util.ErrInternal)
}
