package datagate

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DataGate is a generic interface for the database queries
// T - struct of db table
// F - struct of filter
type DataGate[T, F interface{}] interface {
	GetWithTransaction(tx pgx.Tx) DataGate[T, F]

	Create(ctx context.Context, data T) (int64, error)
	Get(ctx context.Context, filter F) ([]T, error)
	Update(ctx context.Context, filter F, data map[string]interface{}) error
	Delete(ctx context.Context, filter F) error
}

type dataGate[T, F interface{}] struct {
	tableName  string
	primaryKey string
	pool       *pgxpool.Pool
	sqBuilder  squirrel.StatementBuilderType

	tx pgx.Tx
}

// GetWithTransaction runs all queries with the given transaction
func (dg dataGate[T, F]) GetWithTransaction(tx pgx.Tx) DataGate[T, F] {
	return &dataGate[T, F]{
		tableName:  dg.tableName,
		primaryKey: dg.primaryKey,
		pool:       dg.pool,
		sqBuilder:  dg.sqBuilder,

		tx: tx,
	}
}

// Create create a new item in the db
func (dg dataGate[T, F]) Create(ctx context.Context, data T) (int64, error) {
	sql, args, err := dg.sqBuilder.Insert(dg.tableName).
		SetMap(extractStructFieldsByTag(data, insertTag)).
		Suffix("RETURNING " + dg.primaryKey).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var row pgx.Row
	var id int64

	if dg.tx == nil {
		row = dg.pool.QueryRow(ctx, sql, args...)
	} else {
		row = dg.tx.QueryRow(ctx, sql, args...)
	}

	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row: %w", err)
	}

	return id, nil
}

// Get returns a list of items from db
func (dg dataGate[T, F]) Get(ctx context.Context, filter F) ([]T, error) {
	var item T

	var columns []string
	values := extractStructFieldsByTag(item, dbTag)
	for key := range values {
		columns = append(columns, key)
	}

	query := dg.sqBuilder.Select(columns...).From(dg.tableName)
	for _, sqlizer := range buildSqlFiltersFromStruct(filter) {
		query = query.Where(sqlizer)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var items []T
	if dg.tx == nil {
		err = pgxscan.Select(ctx, dg.pool, &items, sql, args...)
	} else {
		err = pgxscan.Select(ctx, dg.tx, &items, sql, args...)
	}

	return items, err
}

// Update updates an item in the db
func (dg dataGate[T, F]) Update(ctx context.Context, filter F, data map[string]interface{}) error {
	query := dg.sqBuilder.Update(dg.tableName).
		SetMap(data)
	for _, sqlizer := range buildSqlFiltersFromStruct(filter) {
		query = query.Where(sqlizer)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	if dg.tx == nil {
		_, err = dg.pool.Exec(ctx, sql, args...)
	} else {
		_, err = dg.tx.Exec(ctx, sql, args...)
	}
	if err != nil {
		return fmt.Errorf("failed exec query: %w", err)
	}

	return nil
}

// Delete deletes an item from the db
func (dg dataGate[T, F]) Delete(ctx context.Context, filter F) error {
	query := dg.sqBuilder.Delete(dg.tableName)
	for _, sqlizer := range buildSqlFiltersFromStruct(filter) {
		query = query.Where(sqlizer)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	if dg.tx == nil {
		_, err = dg.pool.Exec(ctx, sql, args...)
	} else {
		_, err = dg.tx.Exec(ctx, sql, args...)
	}
	if err != nil {
		return fmt.Errorf("failed exec query: %w", err)
	}

	return nil
}

// NewDataGate creates a new DataGate
func NewDataGate[T, F interface{}](
	tableName string,
	primaryKey string,
	pool *pgxpool.Pool,
	sqBuilder squirrel.StatementBuilderType,
) (DataGate[T, F], error) {
	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}

	return &dataGate[T, F]{
		tableName:  tableName,
		primaryKey: primaryKey,
		pool:       pool,
		sqBuilder:  sqBuilder,
	}, nil
}
