package api

import (
	"database/sql"
	"time"
	db "training/11-file-search/db/sqlc"
)

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name" binding:"required"`
	Avatar   string `json:"avatar" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type updateUserRequest struct {
	Email    sql.NullString `json:"email" binding:"required,email"`
	Password sql.NullString `json:"password" binding:"required,min=6"`
	Phone    sql.NullString `json:"phone"`
	FullName sql.NullString `json:"full_name" binding:"required"`
	Avatar   sql.NullString `json:"avatar" binding:"required"`
	Role     sql.NullString `json:"role" binding:"required"`
}

type Meta struct {
	Total  int `json:"metadata"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type Response struct {
	Status int       `json:"status"`
	Data   []db.User `json:"data"`
	Meta   Meta      `json:"meta"`
}
type userResponse struct {
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required"`
}

type DataObject struct {
	AccessToken string `json:"access_token"`
}
type loginUserResponse struct {
	Status int        `json:"status"`
	Error  string     `json:"error"`
	Errors []string   `json:"errors"`
	Data   DataObject `json:"data"`
}

type newDataFile struct {
	FilePath string `json:"filepath"`
	CheckSum string `json:"checksum"`
}
type Metadata struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}
type newFileResponse struct {
	Status int           `json:"status"`
	Data   []newDataFile `json:"data"`
	Meta   Metadata      `json:"meta"`
}
