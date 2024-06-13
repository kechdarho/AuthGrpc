package models

type User struct {
	ID       int64
	Login    string
	Phone    string
	Email    string
	PassHash []byte
	Role     string
}
