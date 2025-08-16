package tests

import (
	"time"
)

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
