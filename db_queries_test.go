package datagate

import (
	"context"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/gippuss/datagate/testutils/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type table struct {
	ID     int64  `db:"id" insert:"id"`
	Name   string `db:"name" insert:"name"`
	Active bool   `db:"active" insert:"active"`
}

type filter struct {
	Name   string `filter:"name"`
	Active bool   `filter:"active"`
}

func TestDataGate_Create(t *testing.T) {
	ctx := context.Background()

	txMock := mocks.NewTxMock(t)
	rowMock := mocks.NewRowMock(t)

	txMock.QueryRowMock.Expect(ctx, "INSERT INTO test_table (active,id,name) VALUES ($1,$2,$3) RETURNING id",
		true, int64(1), "test",
	).Return(rowMock)
	rowMock.ScanMock.Set(func(dest ...any) error {
		if idPtr, ok := dest[0].(*int64); ok {
			*idPtr = 1
		}
		return nil
	})

	dataGate := getDataGate(txMock)
	id, err := dataGate.Create(ctx, table{
		ID:     1,
		Name:   "test",
		Active: true,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), id)
}

func TestDataGate_Update(t *testing.T) {
	ctx := context.Background()

	txMock := mocks.NewTxMock(t)

	txMock.ExecMock.Expect(ctx,
		"UPDATE test_table SET name = $1 WHERE name = $2 AND active = $3",
		"test2", "test", true,
	).Return(pgconn.CommandTag{}, nil)

	dataGate := getDataGate(txMock)
	err := dataGate.Update(ctx, filter{
		Name:   "test",
		Active: true,
	}, map[string]interface{}{
		"name": "test2",
	})

	require.NoError(t, err)
}

func TestDataGate_Delete(t *testing.T) {
	ctx := context.Background()

	txMock := mocks.NewTxMock(t)

	txMock.ExecMock.Expect(ctx,
		"DELETE FROM test_table WHERE name = $1 AND active = $2",
		"test", true,
	).Return(pgconn.CommandTag{}, nil)

	dataGate := getDataGate(txMock)
	err := dataGate.Delete(ctx, filter{
		Name:   "test",
		Active: true,
	})

	require.NoError(t, err)
}

func getDataGate(txMock pgx.Tx) DataGate[table, filter] {
	dataGate, _ := NewDataGate[table, filter](
		"test_table",
		"id",
		&pgxpool.Pool{},
		squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	)
	txDataGate := dataGate.GetWithTransaction(txMock)

	return txDataGate
}
