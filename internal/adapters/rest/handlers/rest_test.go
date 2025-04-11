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
