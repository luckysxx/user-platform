package contract

type GetProfileQuery struct {
	UserID int64
}

type GetProfileResult struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
	UpdatedAt string
}

type UpdateProfileCommand struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
}

type UpdateProfileResult struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
	UpdatedAt string
}
