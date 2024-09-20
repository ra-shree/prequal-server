package utils

import (
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var db *bun.DB

func InitDB() error {
    dsn := "postgres://postgres:postgres@localhost:5432/prequal"
    sqldb, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Printf("Unable to connect to database: %v", err)
        return err
    }

    // Wrap sql.DB with Bun
    db = bun.NewDB(sqldb, pgdialect.New())
    return nil
}


