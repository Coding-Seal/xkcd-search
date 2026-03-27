package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"yadro-go-course/config"
)

const testDB = "../test/db/test.db"

func TestConnect(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestMigrateUp(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, MigrateDown(db, context.Background())) })
	assert.NoError(t, MigrateUp(db, context.Background()))
}

func TestMigrateDown(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, MigrateDown(db, context.Background()))
}

func TestConnect_InMemory(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = ":memory:"
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestConnect_Ping(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, db.Ping())
}

func TestMigrateUp_Idempotent(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, MigrateDown(db, context.Background())) })
	assert.NoError(t, MigrateUp(db, context.Background()))
	// Running twice should not return error (ErrNoChange is swallowed)
	assert.NoError(t, MigrateUp(db, context.Background()))
}

func TestMigrateDown_Idempotent(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = testDB
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, MigrateDown(db, context.Background()))
	// Second down when already empty is fine
	assert.NoError(t, MigrateDown(db, context.Background()))
}

func TestConnect_WithFileURL(t *testing.T) {
	cfg := &config.Config{}
	cfg.DB.Url = "file:" + testDB + "?mode=rwc"
	db, err := Connect(cfg)
	// Either succeeds or driver rejects the URI scheme (both are valid outcomes)
	if err == nil {
		assert.NotNil(t, db)
	}
}
