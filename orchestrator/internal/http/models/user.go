package models

type User struct {
	ID       uint64 `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
