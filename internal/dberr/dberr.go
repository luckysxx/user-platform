package dberr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/luckysxx/user-platform/internal/ent"
)

// 数据库特定错误
var (
	ErrNoRows           = errors.New("记录不存在")
	ErrDuplicateKey     = errors.New("记录已存在")
	ErrForeignKey       = errors.New("关联记录不存在")
	ErrCheckViolation   = errors.New("数据校验失败")
	ErrNotNullViolation = errors.New("必填字段缺失")
	ErrInvalidData      = errors.New("数据格式错误")
	ErrConnection       = errors.New("数据库连接失败")
	ErrConstraint       = errors.New("违反数据约束") // 新增：用于兜底未知的约束错误

	// 用户相关错误
	ErrUsernameDuplicate = errors.New("用户名已存在")
	ErrEmailDuplicate    = errors.New("邮箱已存在")

	// 粘贴相关错误
	ErrShortLinkDuplicate = errors.New("短链接已存在")
)

// ParseDBError 解析数据库错误并转换为业务错误
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	// 1. 记录不存在
	if ent.IsNotFound(err) {
		return ErrNoRows
	}

	// 2. 检查是否是 ent 内存层面的校验错误
	if ent.IsValidationError(err) {
		return ErrInvalidData
	}

	// 3. 约束错误 (数据库层面抛出的)
	if ent.IsConstraintError(err) {
		if mappedErr := parseEntConstraintError(err); mappedErr != nil {
			return mappedErr
		}
		// 如果匹配不到具体的错误名，返回一个通用的约束错误，而不是写死 DuplicateKey
		return ErrConstraint
	}

	// 4. 处理系统级/网络级错误 (不再包在 Constraint 里面)
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "dial tcp") {
		return ErrConnection
	}

	// 5. 其他数据库错误原样返回
	return err
}

// parseEntConstraintError 尝试从 ent 约束错误文本中提取具体业务错误。
func parseEntConstraintError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "users_username_key"):
		return ErrUsernameDuplicate
	case strings.Contains(msg, "users_email_key"):
		return ErrEmailDuplicate
	case strings.Contains(msg, "pastes_short_link_key"):
		return ErrShortLinkDuplicate
	case strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "already exists"):
		return ErrDuplicateKey
	case strings.Contains(msg, "foreign key"):
		return ErrForeignKey
	case strings.Contains(msg, "not-null") || strings.Contains(msg, "not null"):
		return ErrNotNullViolation
	case strings.Contains(msg, "check constraint"):
		return ErrCheckViolation
	default:
		return nil
	}
}

// IsDuplicateKeyError 检查是否是唯一键冲突错误
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrDuplicateKey) ||
		errors.Is(err, ErrUsernameDuplicate) ||
		errors.Is(err, ErrEmailDuplicate) ||
		errors.Is(err, ErrShortLinkDuplicate) {
		return true
	}

	if ent.IsConstraintError(err) {
		mappedErr := parseEntConstraintError(err)
		return mappedErr != nil && IsDuplicateKeyError(mappedErr)
	}
	return false
}

// IsForeignKeyError 检查是否是外键约束错误
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrForeignKey) {
		return true
	}

	if ent.IsConstraintError(err) {
		mappedErr := parseEntConstraintError(err)
		return mappedErr != nil && IsForeignKeyError(mappedErr)
	}
	return false
}

// IsNotFoundError 检查是否是记录不存在错误
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRows) || ent.IsNotFound(err)
}

// IsConnectionError 检查是否是连接错误
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrConnection) {
		return true
	}

	mappedErr := parseEntConstraintError(err)
	if mappedErr != nil {
		return errors.Is(mappedErr, ErrConnection)
	}
	return false
}

// WrapDBError 包装数据库错误，保留原始错误信息
// 适用于需要在错误链中添加上下文的场景
func WrapDBError(err error, operation string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
