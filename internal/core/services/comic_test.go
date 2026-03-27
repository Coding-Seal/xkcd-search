package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
)

func TestNewComicService(t *testing.T) {
	assert.NotNil(t, NewComicService(newComicRepoMock()))
}

func TestComicService_Comic_Found(t *testing.T) {
	repo := newComicRepoMock()
	expected := models.Comic{ID: 7, Title: "Test", ImgURL: "http://example.com/7.png"}
	repo.On("Comic", 7).Return(expected, nil)
	svc := NewComicService(repo)
	got, err := svc.Comic(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestComicService_Comic_NotFound(t *testing.T) {
	repo := newComicRepoMock()
	repo.On("Comic", 999).Return(models.Comic{}, ports.ErrNotFound)
	svc := NewComicService(repo)
	_, err := svc.Comic(context.Background(), 999)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestComicService_Store(t *testing.T) {
	repo := newComicRepoMock()
	comic := models.Comic{ID: 1, Title: "Store Test"}
	repo.On("Store", comic).Return(nil)
	svc := NewComicService(repo)
	err := svc.Store(context.Background(), comic)
	assert.NoError(t, err)
}

func TestComicService_ComicsAll(t *testing.T) {
	repo := newComicRepoMock()
	all := []models.Comic{{ID: 1}, {ID: 2}}
	repo.On("ComicsAll").Return(all, nil)
	svc := NewComicService(repo)
	got, err := svc.ComicsAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, got, 2)
}
