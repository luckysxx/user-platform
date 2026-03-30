package contract

// GetProfileQuery 表示查询用户资料时的输入参数。
type GetProfileQuery struct {
	UserID int64
}

// GetProfileResult 表示查询用户资料时的返回结果。
type GetProfileResult struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
	Birthday  string
	UpdatedAt string
}

// UpdateProfileCommand 表示更新用户资料时的输入参数。
type UpdateProfileCommand struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
	Birthday  string
}

// UpdateProfileResult 表示更新用户资料后的返回结果。
type UpdateProfileResult struct {
	UserID    int64
	Nickname  string
	AvatarURL string
	Bio       string
	Birthday  string
	UpdatedAt string
}
