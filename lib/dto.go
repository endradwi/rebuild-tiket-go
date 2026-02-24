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
	UsedAt *time.Time
	CreatedAt time.Time
}

type ResetPasswordRequest struct {
	Token string `json:"token" form:"token"`
	Password string `json:"password" form:"password"`
}

type UserProfile struct {
	Id int `json:"id"`
	Email string `json:"email"`
	FirstName *string `json:"first_name"`
	LastName *string `json:"last_name"`
	PhoneNumber *int `json:"phone_number"`
	Picture *string `json:"picture"`
	Point *int `json:"point"`
}

type ProfileUpdateRequest struct {
	FirstName *string `form:"first_name"`
	LastName *string `form:"last_name"`
	PhoneNumber *int `form:"phone_number"`
}
