package model

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	// Username: 用户名
	// required: 必填字段
	// min=3: 最少3个字符
	// max=20: 最多20个字符
	// alphanum: 只能包含字母和数字（不能有特殊字符、空格）
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	// Password: 密码
	Password string `json:"password" binding:"required,min=8"`
	// Email: 邮箱
	Email string `json:"email" binding:"required,email"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	// 登录时不需要太严格，只要求必填即可
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
