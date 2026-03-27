package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
)

func TestNewUserService(t *testing.T) {
	assert.NotNil(t, NewUserService(newUserRepoMock()))
}

func TestUserService_UserLogin_Found(t *testing.T) {
	repo := newUserRepoMock()
	expected := models.User{ID: 1, Login: "alice", IsAdmin: false}
	repo.On("UserLogin", "alice").Return(expected, nil)
	svc := NewUserService(repo)
	got, err := svc.UserLogin(context.Background(), "alice")
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestUserService_UserLogin_NotFound(t *testing.T) {
	repo := newUserRepoMock()
	repo.On("UserLogin", "ghost").Return(models.User{}, ports.ErrNotFound)
	svc := NewUserService(repo)
	_, err := svc.UserLogin(context.Background(), "ghost")
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestUserService_AddUser(t *testing.T) {
	repo := newUserRepoMock()
	user := &models.User{Login: "newuser", Password: []byte("hash"), IsAdmin: false}
	repo.On("AddUser", user).Return(nil)
	svc := NewUserService(repo)
	err := svc.AddUser(context.Background(), user)
	assert.NoError(t, err)
}

func TestUserService_UserID(t *testing.T) {
	repo := newUserRepoMock()
	expected := models.User{ID: 5, Login: "bob"}
	repo.On("UserID", int64(5)).Return(expected, nil)
	svc := NewUserService(repo)
	got, err := svc.UserID(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestUserService_RemoveUser(t *testing.T) {
	repo := newUserRepoMock()
	repo.On("RemoveUser", int64(3)).Return(nil)
	svc := NewUserService(repo)
	err := svc.RemoveUser(context.Background(), 3)
	assert.NoError(t, err)
}
