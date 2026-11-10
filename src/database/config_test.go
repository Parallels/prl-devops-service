package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/stretchr/testify/assert"
)

func resetConfig(t *testing.T) {
	t.Helper()
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	_ = config.New(ctx)
}

func TestDefaultDatabaseConfig_UsesDatabaseFolder(t *testing.T) {
	resetConfig(t)
	tmp := t.TempDir()
	t.Setenv("DATABASE_FOLDER", tmp)

	cfg := DefaultDatabaseConfig()
	assert.Equal(t, filepath.Join(tmp, DefaultDatabaseFilename), cfg.Path)
	assert.Equal(t, DBTypeSQLite, cfg.Type)
	assert.True(t, cfg.AutoMigrate)
	assert.Empty(t, cfg.DSN)
}

func TestDefaultDatabaseConfig_UsesHomeWhenFolderUnset(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root defaults to /etc/parallels-devops by design; HOME path covered only for non-root")
	}
	resetConfig(t)
	tmp := t.TempDir()
	t.Setenv("DATABASE_FOLDER", "")
	t.Setenv("HOME", tmp)

	cfg := DefaultDatabaseConfig()
	assert.Equal(t, filepath.Join(tmp, ".parallels-devops", DefaultDatabaseFilename), cfg.Path)
}

func TestDefaultDatabaseConfig_RootUsesEtc(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("root-only branch")
	}
	resetConfig(t)
	t.Setenv("DATABASE_FOLDER", "")

	cfg := DefaultDatabaseConfig()
	assert.Equal(t, filepath.Join("/etc/parallels-devops", DefaultDatabaseFilename), cfg.Path)
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		cfg  DatabaseConfig
		err  error
	}{
		{"empty type defaults to sqlite", DatabaseConfig{Path: "/tmp/x.db"}, nil},
		{"uppercase sqlite accepted", DatabaseConfig{Type: "SQLITE", Path: "/tmp/x.db"}, nil},
		{"mysql unsupported", DatabaseConfig{Type: "mysql", Path: "/tmp/x.db"}, ErrUnsupportedDatabaseType},
		{"postgres blank dsn rejected", DatabaseConfig{Type: "postgres", DSN: " "}, ErrPostgresDSNRequired},
		{"postgres ok", DatabaseConfig{Type: "postgres", DSN: "host=h dbname=d"}, nil},
		{"sqlite blank path rejected", DatabaseConfig{Type: "sqlite", Path: "  "}, ErrSQLitePathRequired},
		{"sqlite memory path ok", DatabaseConfig{Type: "sqlite", Path: MemoryPath}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}
		})
	}
}

func TestPostgresPoolDefaults(t *testing.T) {
	zero := PostgresPoolConfig{}
	assert.Equal(t, DefaultPostgresMaxOpenConns, zero.MaxOpen())
	assert.Equal(t, DefaultPostgresMaxIdleConns, zero.MaxIdle())
	assert.Equal(t, DefaultPostgresConnMaxLifetime, zero.Lifetime())
	assert.Equal(t, time.Duration(0), zero.IdleTime())

	set := PostgresPoolConfig{
		MaxOpenConns:    11,
		MaxIdleConns:    3,
		ConnMaxLifetime: 7 * time.Minute,
		ConnMaxIdleTime: 90 * time.Second,
	}
	assert.Equal(t, 11, set.MaxOpen())
	assert.Equal(t, 3, set.MaxIdle())
	assert.Equal(t, 7*time.Minute, set.Lifetime())
	assert.Equal(t, 90*time.Second, set.IdleTime())
}

func TestFromApplicationConfig_SeamDelegatesToDefault(t *testing.T) {
	resetConfig(t)
	tmp := t.TempDir()
	t.Setenv("DATABASE_FOLDER", tmp)

	assert.Equal(t, DefaultDatabaseConfig(), FromApplicationConfig())
}
