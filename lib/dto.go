package lib

import "time"


type User struct {
	Id int 
	Email string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type UserRole struct {
	User User
	RoleId int `json:"role_id" form:"role_id"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
}

type ResetPassword struct {
	Id int
	ProfileId int
	TokenHash string
	ExpiredAt time.Time
	UsedAt time.Time
	CreatedAt time.Time
}
