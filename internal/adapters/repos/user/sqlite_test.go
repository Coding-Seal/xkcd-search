package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/config"
	"yadro-go-course/db"
	"yadro-go-course/internal/core/models"
	"yadro-go-course/internal/core/ports"
)

func newTestRepo(t *testing.T) *SqliteRepo {
	t.Helper()
	cfg := &config.Config{DB: config.DB{Url: ":memory:"}}
	conn, err := db.Connect(cfg)
	assert.NoError(t, err)
	ctx := context.Background()
	assert.NoError(t, db.MigrateUp(conn, ctx))
	t.Cleanup(func() { assert.NoError(t, db.MigrateDown(conn, ctx)) })
	return NewSqliteRepo(conn)
}

func TestSqliteRepo_UserID(t *testing.T) {
	testUser := models.User{Login: "Bob"}
	ctx := context.Background()
	repo := newTestRepo(t)

	assert.NoError(t, repo.AddUser(ctx, &testUser))
	u, err := repo.UserID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, testUser, u)

	_, err = repo.UserID(ctx, 30)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteRepo_UserLogin(t *testing.T) {
	testUser := models.User{Login: "Bob"}
	ctx := context.Background()
	repo := newTestRepo(t)

	assert.NoError(t, repo.AddUser(ctx, &testUser))

	u, err := repo.UserLogin(ctx, testUser.Login)
	assert.NoError(t, err)
	assert.Equal(t, testUser, u)

	_, err = repo.UserLogin(ctx, "Alice")
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteRepo_RemoveUser(t *testing.T) {
	testUser := models.User{Login: "Bob"}
	ctx := context.Background()
	repo := newTestRepo(t)

	assert.NoError(t, repo.AddUser(ctx, &testUser))
	u, err := repo.UserID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, testUser, u)

	assert.NoError(t, repo.RemoveUser(ctx, testUser.ID))
	_, err = repo.UserID(ctx, testUser.ID)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteRepo_AddUser_AssignsID(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)
	u := &models.User{Login: "Charlie"}
	assert.NoError(t, repo.AddUser(ctx, u))
	assert.NotZero(t, u.ID, "AddUser should populate the ID field")
}

func TestSqliteRepo_UserID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)
	_, err := repo.UserID(ctx, 999)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteRepo_UserLogin_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)
	_, err := repo.UserLogin(ctx, "ghost")
	assert.ErrorIs(t, err, ports.ErrNotFound)
}

func TestSqliteRepo_MultipleUsers(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	users := []*models.User{
		{Login: "alice"},
		{Login: "bob"},
		{Login: "carol"},
	}
	for _, u := range users {
		assert.NoError(t, repo.AddUser(ctx, u))
	}

	for _, u := range users {
		got, err := repo.UserLogin(ctx, u.Login)
		assert.NoError(t, err)
		assert.Equal(t, u.Login, got.Login)
	}
}

func TestSqliteRepo_RemoveUser_ThenNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)
	u := &models.User{Login: "temp"}
	assert.NoError(t, repo.AddUser(ctx, u))
	assert.NoError(t, repo.RemoveUser(ctx, u.ID))
	_, err := repo.UserLogin(ctx, u.Login)
	assert.ErrorIs(t, err, ports.ErrNotFound)
}
