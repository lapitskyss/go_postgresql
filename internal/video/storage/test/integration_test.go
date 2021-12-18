//go:build integration_tests
// +build integration_tests

package test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"

	"github.com/lapitskyss/go_postgresql/internal/store"
	"github.com/lapitskyss/go_postgresql/internal/video/storage"
)

const (
	DB_HOST     = "127.0.0.1"
	DB_USER     = "gopher"
	DB_PASSWORD = "password"
	DB_NAME     = "gopher_youtube"
)

var DB_PORT = ""

var DBStore *store.Store

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	setupResult, err := setup()
	if err != nil {
		log.Println("setup err: ", err)
		return -1
	}
	defer teardown(setupResult)
	return m.Run()
}

type teardownPack struct {
	OldEnvVars map[string]string
}

const dataDir = "data"

type setupResult struct {
	Pool              *dockertest.Pool
	PostgresContainer *dockertest.Resource
}

const dockerMaxWait = time.Second * 5

func setup() (r *setupResult, err error) {
	testFileDir, err := getTestFileDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get the script dir: %w", err)
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docketest pool: %w", err)
	}
	pool.MaxWait = dockerMaxWait

	postgresContainer, err := runPostgresContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the Postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			if err := pool.Purge(postgresContainer); err != nil {
				log.Println("failed to purge the postgres container: %w", err)
			}
		}
	}()

	migrationContainer, err := runMigrationContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the migration container: %w", err)
	}

	defer func() {
		if err := pool.Purge(migrationContainer); err != nil {
			err = fmt.Errorf("failed to purge the migration container: %w", err)
		}
	}()

	if err := pool.Retry(func() error {
		err := prepopulateDB(testFileDir)
		if err != nil {
			log.Printf("populate DB err: %v", err)
		}
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to prepopulate the DB: %w", err)
	}

	return &setupResult{
		Pool:              pool,
		PostgresContainer: postgresContainer,
	}, nil
}

func getTestFileDir() (string, error) {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get the caller info")
	}
	fileDir := filepath.Dir(fileName)
	dir, err := filepath.Abs(fileDir)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path to the directory %s: %w", dir, err)
	}
	return fileDir, nil
}

func runPostgresContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	postgresContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "14.0",
			Env: []string{
				"POSTGRES_PASSWORD=password",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/docker-entrypoint-initdb.d",
					Source: filepath.Join(testFileDir, "init"),
					Type:   "bind",
				},
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the postgres docker container: %w", err)
	}
	postgresContainer.Expire(120)

	DB_PORT = postgresContainer.GetPort("5432/tcp")

	// Wait for the DB to start
	if err := pool.Retry(func() error {
		db, err := getDBConnector()
		if err != nil {
			return fmt.Errorf("failed to get a DB connector: %w", err)
		}
		return db.Ping(context.Background())
	}); err != nil {
		pool.Purge(postgresContainer)
		return nil, fmt.Errorf("failed to ping the created DB: %w", err)
	}
	return postgresContainer, nil
}

func runMigrationContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	migrationsDir, err := filepath.Abs(filepath.Join(testFileDir, "../../../../migrations"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute path of the migrations dir: %w", err)
	}
	migrationContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "migrate/migrate",
			Tag:        "v4.15.0",
			Cmd: []string{
				"-path=/migrations",
				fmt.Sprintf(
					"-database=%s",
					composeConnectionString(),
				),
				"up",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/migrations",
					Source: migrationsDir,
					Type:   "bind",
				},
			}
			config.NetworkMode = "host"
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the migration container: %w", err)
	}

	return migrationContainer, err
}

func prepopulateDB(testFileDir string) error {
	prepopulateScriptPath := filepath.Join(testFileDir, "prepopulate_db.sql")
	scriptBytes, err := os.ReadFile(prepopulateScriptPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", prepopulateScriptPath, err)
	}
	conn, err := getDBConnector()
	if err != nil {
		return fmt.Errorf("failed to get a DB connector: %w", err)
	}
	if _, err := conn.Exec(context.Background(), string(scriptBytes)); err != nil {
		return fmt.Errorf("failed to execute the prepopulate script: %w", err)
	}
	return nil
}

func teardown(r *setupResult) {
	if err := r.Pool.Purge(r.PostgresContainer); err != nil {
		log.Printf("failed to purge the Postgres container: %v", err)
	}
}

func TestFindVideosByTitle(t *testing.T) {
	videoStorage := storage.NewVideoStorage(DBStore.Connection(), mocktracer.New())

	videos, err := videoStorage.FindVideosByTitle(context.Background(), "go")
	if err != nil {
		t.Fatalf("failed to get a connector to the DB: %v", err)
	}

	if len(videos) != 1 {
		t.Fatalf("expected single query result")
	}

	if videos[0].Title != "Golang awesome" {
		t.Fatalf("incorrect video title returned, expected 'Golang awesome'")
	}
}

func getDBConnector() (*pgxpool.Pool, error) {
	if DBStore == nil {
		pgStore, err := store.Connect(context.Background(), composeConnectionString(), zap.NewNop())
		if err != nil {
			return nil, err
		}

		DBStore = pgStore
	}

	return DBStore.Connection(), nil
}

func composeConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", DB_USER, url.QueryEscape(DB_PASSWORD), DB_HOST, DB_PORT, DB_NAME)
}
