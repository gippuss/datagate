# Datagate

**Datagate** is a simple and type-safe Go library for CRUD operations with PostgreSQL. It is built on top of the [pgx](https://github.com/jackc/pgx) driver and uses Go generics to reduce boilerplate. You describe your models and filters as Go structs, and Datagate handles SQL generation for you.

## Features

- Generic CRUD interface for any struct
- Works directly with [pgx](https://github.com/jackc/pgx) connection pool and transactions
- Automatic mapping of struct fields to table columns using tags
- Flexible filtering with filter structs
- Easy integration with [squirrel](https://github.com/Masterminds/squirrel) for SQL building

## Installation

```bash
go get github.com/gippuss/datagate
```

## Quick Start

### 1. Define your models

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name" insert:"name"`
    Email     string    `db:"email" insert:"email"`
    Age       int       `db:"age" insert:"age"`
    CreatedAt time.Time `db:"created_at" insert:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

type UserFilter struct {
    ID    *int64  `filter:"id"`
    Name  *string `filter:"name"`
    Email *string `filter:"email"`
    Age   *int    `filter:"age"`
}
```
- Use `db` tag for reading from the database
- Use `insert` tag for inserting new records
- Use `filter` tag for filtering in queries

### 2. Initialize Datagate

```go
import (
    "github.com/gippuss/datagate"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/Masterminds/squirrel"
)

pool, _ := pgxpool.New(ctx, "postgres://user:pass@host:port/dbname")
sqBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

dataGate, err := datagate.NewDataGate[User, UserFilter](
    "users",      // table name
    "id",         // primary key
    pool,
    sqBuilder,
)
```

### 3. CRUD operations

#### Create
```go
id, err := dataGate.Create(ctx, User{Name: "Alice", Email: "alice@example.com", Age: 25})
```

#### Read
```go
users, err := dataGate.Get(ctx, UserFilter{Name: &name})
```

#### Update
```go
err := dataGate.Update(ctx, UserFilter{ID: &id}, map[string]interface{}{"name": "Bob"})
```

#### Delete
```go
err := dataGate.Delete(ctx, UserFilter{ID: &id})
```

#### Transactions
```go
tx, _ := pool.Begin(ctx)
txDataGate := dataGate.GetWithTransaction(tx)
id, err := txDataGate.Create(ctx, user)
// ...
tx.Commit(ctx)
```
