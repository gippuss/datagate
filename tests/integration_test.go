package tests

import (
	"context"
	"testing"
	"time"

	"github.com/gippuss/datagate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func (s *DataGateIntegrationTestSuite) TestCreate() {
	ctx := context.Background()

	user := User{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
	}

	id, err := s.dataGate.Create(ctx, user)

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), id, int64(1))

	// Проверяем, что запись действительно создана
	filter := UserFilter{ID: &id}
	users, err := s.dataGate.Get(ctx, filter)

	require.NoError(s.T(), err)
	require.Len(s.T(), users, 1)
	assert.Equal(s.T(), "John Doe", users[0].Name)
	assert.Equal(s.T(), "john@example.com", users[0].Email)
	assert.Equal(s.T(), 30, users[0].Age)
}

func (s *DataGateIntegrationTestSuite) TestGet() {
	ctx := context.Background()

	users := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 25, CreatedAt: time.Now()},
		{Name: "Bob", Email: "bob@example.com", Age: 30, CreatedAt: time.Now()},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, CreatedAt: time.Now()},
	}

	var ids []int64
	for _, user := range users {
		id, err := s.dataGate.Create(ctx, user)
		require.NoError(s.T(), err)
		ids = append(ids, id)
	}

	allUsers, err := s.dataGate.Get(ctx, UserFilter{})
	assert.NoError(s.T(), err)
	assert.Len(s.T(), allUsers, 3)

	filter := UserFilter{ID: &ids[0]}
	result, err := s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), "Alice", result[0].Name)

	name := "Bob"
	nameFilter := UserFilter{Name: &name}
	result, err = s.dataGate.Get(ctx, nameFilter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), "Bob", result[0].Name)

	age := 35
	ageFilter := UserFilter{Age: &age}
	result, err = s.dataGate.Get(ctx, ageFilter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result, 1)
	assert.Equal(s.T(), "Charlie", result[0].Name)
}

func (s *DataGateIntegrationTestSuite) TestUpdate() {
	ctx := context.Background()

	user := User{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
	}

	id, err := s.dataGate.Create(ctx, user)
	require.NoError(s.T(), err)

	updateData := map[string]interface{}{
		"name": "John Smith",
		"age":  31,
	}

	filter := UserFilter{ID: &id}
	err = s.dataGate.Update(ctx, filter, updateData)
	assert.NoError(s.T(), err)

	updatedUsers, err := s.dataGate.Get(ctx, filter)
	require.NoError(s.T(), err)
	require.Len(s.T(), updatedUsers, 1)

	assert.Equal(s.T(), "John Smith", updatedUsers[0].Name)
	assert.Equal(s.T(), 31, updatedUsers[0].Age)
	assert.Equal(s.T(), "john@example.com", updatedUsers[0].Email)
}

func (s *DataGateIntegrationTestSuite) TestDelete() {
	ctx := context.Background()

	user1 := User{Name: "Alice", Email: "alice@example.com", Age: 25, CreatedAt: time.Now()}
	user2 := User{Name: "Bob", Email: "bob@example.com", Age: 30, CreatedAt: time.Now()}

	id1, err := s.dataGate.Create(ctx, user1)
	require.NoError(s.T(), err)
	id2, err := s.dataGate.Create(ctx, user2)
	require.NoError(s.T(), err)

	filter := UserFilter{ID: &id1}
	err = s.dataGate.Delete(ctx, filter)
	assert.NoError(s.T(), err)

	deletedUsers, err := s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), deletedUsers, 0)

	filter2 := UserFilter{ID: &id2}
	remainingUsers, err := s.dataGate.Get(ctx, filter2)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), remainingUsers, 1)
}

func (s *DataGateIntegrationTestSuite) TestTransactionCommit() {
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	require.NoError(s.T(), err)

	txDataGate := s.dataGate.GetWithTransaction(tx)

	user := User{
		Name:      "Transaction User",
		Email:     "tx@example.com",
		Age:       25,
		CreatedAt: time.Now(),
	}

	id, err := txDataGate.Create(ctx, user)
	require.NoError(s.T(), err)

	filter := UserFilter{ID: &id}
	users, err := s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)

	err = tx.Commit(ctx)
	require.NoError(s.T(), err)

	users, err = s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 1)
}

func (s *DataGateIntegrationTestSuite) TestTransactionRollback() {
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	require.NoError(s.T(), err)

	txDataGate := s.dataGate.GetWithTransaction(tx)

	user := User{
		Name:      "Rollback User",
		Email:     "rollback@example.com",
		Age:       25,
		CreatedAt: time.Now(),
	}

	id, err := txDataGate.Create(ctx, user)
	require.NoError(s.T(), err)

	err = tx.Rollback(ctx)
	require.NoError(s.T(), err)

	filter := UserFilter{ID: &id}
	users, err := s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)
}

// Тесты граничных случаев и ошибок

func (s *DataGateIntegrationTestSuite) TestCreateDuplicateEmail() {
	ctx := context.Background()

	user1 := User{Name: "User1", Email: "duplicate@example.com", Age: 25, CreatedAt: time.Now()}
	user2 := User{Name: "User2", Email: "duplicate@example.com", Age: 30, CreatedAt: time.Now()}

	_, err := s.dataGate.Create(ctx, user1)
	require.NoError(s.T(), err)

	// Попытка создать пользователя с дублирующимся email
	_, err = s.dataGate.Create(ctx, user2)
	assert.Error(s.T(), err)
}

func (s *DataGateIntegrationTestSuite) TestGetNonExistent() {
	ctx := context.Background()

	nonExistentID := int64(99999)
	filter := UserFilter{ID: &nonExistentID}

	users, err := s.dataGate.Get(ctx, filter)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 0)
}

func (s *DataGateIntegrationTestSuite) TestUpdateNonExistent() {
	ctx := context.Background()

	nonExistentID := int64(99999)
	filter := UserFilter{ID: &nonExistentID}
	updateData := map[string]interface{}{"name": "Updated"}

	err := s.dataGate.Update(ctx, filter, updateData)
	assert.NoError(s.T(), err) // Update не возвращает ошибку, если ничего не обновлено
}

func (s *DataGateIntegrationTestSuite) TestDeleteNonExistent() {
	ctx := context.Background()

	nonExistentID := int64(99999)
	filter := UserFilter{ID: &nonExistentID}

	err := s.dataGate.Delete(ctx, filter)
	assert.NoError(s.T(), err) // Delete не возвращает ошибку, если ничего не удалено
}

func (s *DataGateIntegrationTestSuite) TestNewDataGateValidation() {
	// Тест с nil pool
	_, err := datagate.NewDataGate[User, UserFilter]("users", "id", nil, s.sqBuilder)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "pool is nil")
}

func TestDataGateIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DataGateIntegrationTestSuite))
}
