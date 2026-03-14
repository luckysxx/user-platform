package dto

// RegisterRequest is the HTTP DTO for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the HTTP DTO for user login.
type LoginRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// RegisterResponse is the HTTP DTO for registration response.
type RegisterResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// LoginResponse is the HTTP DTO for login response.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
