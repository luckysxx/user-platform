package contract

// RegisterCommand is the service-layer input for registration.
type RegisterCommand struct {
	Username string
	Password string
}

// RegisterResult is the service-layer output for registration.
type RegisterResult struct {
	UserID   int64
	Username string
}
