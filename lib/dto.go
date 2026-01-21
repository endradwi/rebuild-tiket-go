package lib


type User struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Result  any    `json:"result"`
}
