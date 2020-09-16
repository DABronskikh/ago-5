package dto

import "time"

type CardDTO struct {
	Id      int64     `json:"id"`
	Number  string    `json:"number"`
	Balance int64     `json:"balance"`
	Issuer  string    `json:"issuer"`
	Holder  string    `json:"holder"`
	UserId  int64     `json:"user_id"`
	Status  string    `json:"status"`
	Created time.Time `json:"created"`
}
