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
