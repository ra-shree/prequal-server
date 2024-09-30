package models

import (
	"context"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
)

var db *bun.DB
var ctx = context.Background()

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID        int64     `json:"id" bun:"id,pk,autoincrement"`
	Username  string    `json:"username" bun:"username,unique,notnull"`
	Email     string    `json:"email" bun:"email,unique,notnull"`
	Password  string    `json:"password" bun:"password,notnull"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,default:current_timestamp"`
}

func InsertUser(user User) error {
	_, err := db.NewInsert().Model(&user).Exec(ctx)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return err
	}
	return nil
}

func GetUsersinfo() ([]User, error) {
	var users []User
	err := db.NewSelect().Model(&users).Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
