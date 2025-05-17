package model

type PublicUser struct {
	UserId     string     `json:"userId"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Avatar     string     `json:"avatar"`
	Background string     `json:"background"`
	Status     UserStatus `json:"status"`
}
