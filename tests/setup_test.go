package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gippuss/datagate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DataGateIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	pool      *pgxpool.Pool
	dataGate  datagate.DataGate[User, UserFilter]
	sqBuilder squirrel.StatementBuilderType
}

func (s *DataGateIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(s.T(), err)
	s.container = container

	host, err := container.Host(ctx)
	require.NoError(s.T(), err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, port.Port())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(s.T(), err)
	s.pool = pool

	err = pool.Ping(ctx)
	require.NoError(s.T(), err)

	s.createTestTable()

	s.sqBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	dataGate, err := datagate.NewDataGate[User, UserFilter](
		"users",
		"id",
		pool,
		s.sqBuilder,
	)
	require.NoError(s.T(), err)
	s.dataGate = dataGate
}

func (s *DataGateIntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		s.container.Terminate(context.Background())
	}
}

func (s *DataGateIntegrationTestSuite) SetupTest() {
	s.clearTestData()
}

func (s *DataGateIntegrationTestSuite) createTestTable() {
	ctx := context.Background()

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			age INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`

	_, err := s.pool.Exec(ctx, createTableSQL)
	require.NoError(s.T(), err)
}

func (s *DataGateIntegrationTestSuite) clearTestData() {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, "DELETE FROM users")
	require.NoError(s.T(), err)
}
