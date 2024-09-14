package models

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var db *bun.DB
var ctx = context.Background()

func InitDB(dsn string) error {
    
    sqldb, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Printf("Unable to connect to database: %v", err)
        return err
    }

    // Wrap sql.DB with Bun
    db = bun.NewDB(sqldb, pgdialect.New())
    return nil
}

func InsertUser(user User) error {
    _, err := db.NewInsert().Model(&user).Exec(ctx)
    if err != nil {
        log.Printf("Error inserting user: %v", err)
        return err
    }
    return nil
}

func GetUsers() ([]User, error) {
    var users []User
    err := db.NewSelect().Model(&users).Order("id ASC").Scan(ctx)
    if err != nil {
        return nil, err
    }
    return users, nil
}
