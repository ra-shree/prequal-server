package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
    bun.BaseModel `bun:"table:users"` 

    ID        int64     `json:"id" bun:"id,pk,autoincrement"`
    Username  string    `json:"username" bun:"username,unique,notnull"`
    Email     string    `json:"email" bun:"email,unique,notnull"`
    Password  string    `json:"password" bun:"password,notnull"`
    CreatedAt time.Time `json:"created_at" bun:"created_at,default:current_timestamp"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}