package tests

import (
	"context"
	"fmt"
	"os"

	"github.com/gippuss/datagate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DataGateIntegrationTestSuite struct {
	suite.Suite
	pool     *pgxpool.Pool
	dataGate datagate.DataGate[User, UserFilter]
}

func (s *DataGateIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	dbName := getEnvOrDefault("TEST_DB_NAME", "testdb")
	dbUser := getEnvOrDefault("TEST_DB_USER", "testuser")
	dbPassword := getEnvOrDefault("TEST_DB_PASSWORD", "testpass")

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable",
		dbUser, dbPassword, dbName)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(s.T(), err)
	s.pool = pool

	err = pool.Ping(ctx)
	require.NoError(s.T(), err)

	s.createTestTable()

	dataGate, err := datagate.NewDataGate[User, UserFilter](
		"users",
		"id",
		pool,
	)
	require.NoError(s.T(), err)
	s.dataGate = dataGate
}

func (s *DataGateIntegrationTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
