package contract

// RegisterCommand is the service-layer input for registration.
type RegisterCommand struct {
	Email    string
	Username string
	Password string
}

// RegisterResult is the service-layer output for registration.
type RegisterResult struct {
	Email    string
	UserID   int64
	Username string
}
