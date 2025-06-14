package testutils

import "github.com/jackc/pgx/v5"

//go:generate minimock -i Tx -o mocks/tx_mock.go
type Tx interface {
	pgx.Tx
}

//go:generate minimock -i Row -o mocks/row_mock.go
type Row interface {
	pgx.Row
}

//go:generate minimock -i Rows -o mocks/rows_mock.go
type Rows interface {
	pgx.Rows
}
