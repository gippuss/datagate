package datagate

import (
	"context"
	"errors"
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

func newTestDataGate(t *testing.T, txMock pgx.Tx) DataGate[table, filter] {
	t.Helper()

	dataGate, _ := NewDataGate[table, filter](
		"test_table",
		"id",
		&pgxpool.Pool{},
		squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	)
	return dataGate.GetWithTransaction(txMock)
}

func TestDataGate_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   table
		mockErr error
	}{
		{
			name:  "success",
			input: table{ID: 1, Name: "test", Active: true},
		},
		{
			name:    "scan error",
			input:   table{ID: 2, Name: "fail", Active: false},
			mockErr: errors.New("scan error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			txMock := mocks.NewTxMock(t)
			rowMock := mocks.NewRowMock(t)

			txMock.QueryRowMock.Expect(ctx, "INSERT INTO test_table (active,id,name) VALUES ($1,$2,$3) RETURNING id",
				tt.input.Active, tt.input.ID, tt.input.Name,
			).Return(rowMock)
			rowMock.ScanMock.Set(func(dest ...any) error {
				if tt.mockErr != nil {
					return tt.mockErr
				}
				if idPtr, ok := dest[0].(*int64); ok {
					*idPtr = 1
				}
				return nil
			})

			dataGate := newTestDataGate(t, txMock)
			id, err := dataGate.Create(ctx, tt.input)

			if tt.mockErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, int64(1), id)
			}
		})
	}
}

// func TestDataGate_Get(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name    string
// 		filter  filter
// 		mockErr error
// 	}{
// 		{
// 			name:   "success",
// 			filter: filter{Name: "test", Active: true},
// 		},
// 		{
// 			name:    "scan error",
// 			filter:  filter{Name: "fail", Active: false},
// 			mockErr: errors.New("scan error"),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			ctx := context.Background()
// 			txMock := mocks.NewTxMock(t)
// 			rowsMock := mocks.NewRowsMock(t)

// 			txMock.QueryMock.Expect(ctx, "SELECT id, name, active FROM test_table WHERE name = $1 AND active = $2",
// 				tt.filter.Name, tt.filter.Active,
// 			).Return(rowsMock, tt.mockErr)
// 			rowsMock.NextMock.Expect().Return(false)
// 			rowsMock.ErrMock.Expect().Return(tt.mockErr)
// 			rowsMock.CloseMock.Expect().Return()

// 			dataGate := newTestDataGate(t, txMock)
// 			id, err := dataGate.Get(ctx, tt.filter)

// 			if tt.mockErr != nil {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, []table{nil}, id)
// 			}
// 		})
// 	}
// }

func TestDataGate_Update(t *testing.T) {
	t.Parallel()

	type args struct {
		flt filter
		upd map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		mockErr error
	}{
		{
			name: "success",
			args: args{
				flt: filter{Name: "test", Active: true},
				upd: map[string]interface{}{"name": "test2"},
			},
		},
		{
			name: "exec error",
			args: args{
				flt: filter{Name: "fail", Active: false},
				upd: map[string]interface{}{"name": "fail2"},
			},
			mockErr: errors.New("exec error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			txMock := mocks.NewTxMock(t)
			txMock.ExecMock.Expect(ctx,
				"UPDATE test_table SET name = $1 WHERE name = $2 AND active = $3",
				tt.args.upd["name"], tt.args.flt.Name, tt.args.flt.Active,
			).Return(pgconn.CommandTag{}, tt.mockErr)

			dataGate := newTestDataGate(t, txMock)
			err := dataGate.Update(ctx, tt.args.flt, tt.args.upd)

			if tt.mockErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDataGate_Delete(t *testing.T) {
	t.Parallel()

	type args struct {
		flt filter
	}
	tests := []struct {
		name    string
		args    args
		mockErr error
	}{
		{
			name: "success",
			args: args{flt: filter{Name: "test", Active: true}},
		},
		{
			name:    "delete error",
			args:    args{flt: filter{Name: "fail", Active: false}},
			mockErr: errors.New("delete error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			txMock := mocks.NewTxMock(t)
			txMock.ExecMock.Expect(ctx,
				"DELETE FROM test_table WHERE name = $1 AND active = $2",
				tt.args.flt.Name, tt.args.flt.Active,
			).Return(pgconn.CommandTag{}, tt.mockErr)

			dataGate := newTestDataGate(t, txMock)
			err := dataGate.Delete(ctx, tt.args.flt)

			if tt.mockErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
