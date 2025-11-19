package auth

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type MeResponse struct {
	User *User `json:"user"`
}
