package model

type SearchUser struct {
	UserId      string     `json:"userId"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Avatar      string     `json:"avatar"`
	Background  string     `json:"background"`
	Status      UserStatus `json:"status"`
	IsFollowed  bool       `json:"isFollowed"`
	IsFollowing bool       `json:"isFollowing"`
	IsBlocked   bool       `json:"isBlocked"`
}
