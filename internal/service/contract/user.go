package contract

// RegisterCommand is the service-layer input for registration.
type RegisterCommand struct {
	Username string
	Password string
	Email    string
}

// LoginCommand is the service-layer input for login.
type LoginCommand struct {
	Username string
	Password string
}

// RegisterResult is the service-layer output for registration.
type RegisterResult struct {
	UserID   int64
	Username string
	Email    string
}

// LoginResult is the service-layer output for login.
type LoginResult struct {
	Token    string
	UserID   int64
	Username string
	Email    string
}
