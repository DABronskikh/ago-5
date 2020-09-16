package dto

type UserDTO struct {
	Id       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Roles    string `json:"roles"`
}