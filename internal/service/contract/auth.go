package contract

// LoginCommand is the service-layer input for login.
type LoginCommand struct {
	Username string
	Password string
	AppCode  string
}

// LoginResult is the service-layer output for login.
type LoginResult struct {
	AccessToken  string
	RefreshToken string
	UserID       int64
	Username     string
}

// VerifyTokenCommand is the service-layer input for access token verification.
type VerifyTokenCommand struct {
	Token string
}

// VerifyTokenResult is the service-layer output for access token verification.
type VerifyTokenResult struct {
	UserID   int64
	Username string
}

// RefreshTokenCommand is the service-layer input for refresh token rotation.
type RefreshTokenCommand struct {
	Token string
}

// RefreshTokenResult is the service-layer output for refresh token rotation.
type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}
