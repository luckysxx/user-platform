package dberr

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// 数据库特定错误
var (
	ErrNoRows           = sql.ErrNoRows
	ErrDuplicateKey     = errors.New("记录已存在")
	ErrForeignKey       = errors.New("关联记录不存在")
	ErrCheckViolation   = errors.New("数据校验失败")
	ErrNotNullViolation = errors.New("必填字段缺失")
	ErrInvalidData      = errors.New("数据格式错误")
	ErrConnection       = errors.New("数据库连接失败")

	// 用户相关错误
	ErrUsernameDuplicate = errors.New("用户名已存在")
	ErrEmailDuplicate    = errors.New("邮箱已存在")

	// 粘贴相关错误
	ErrShortLinkDuplicate = errors.New("短链接已存在")
)

// PostgreSQL 错误代码
// 参考: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	// 完整性约束违反
	PgErrUniqueViolation     = "23505" // 唯一键冲突
	PgErrForeignKeyViolation = "23503" // 外键约束违反
	PgErrCheckViolation      = "23514" // CHECK 约束违反
	PgErrNotNullViolation    = "23502" // NOT NULL 约束违反

	// 数据异常
	PgErrInvalidTextRepresentation = "22P02" // 无效的文本表示
	PgErrNumericValueOutOfRange    = "22003" // 数值超出范围
	PgErrDivisionByZero            = "22012" // 除零错误

	// 语法错误
	PgErrSyntaxError     = "42601" // 语法错误
	PgErrUndefinedColumn = "42703" // 未定义的列
	PgErrUndefinedTable  = "42P01" // 未定义的表

	// 连接异常
	PgErrConnectionException    = "08000" // 连接异常
	PgErrConnectionDoesNotExist = "08003" // 连接不存在
	PgErrConnectionFailure      = "08006" // 连接失败
)

// ParseDBError 解析数据库错误并转换为业务错误
// 返回标准 Go error，由 Handler 层负责转换为 CustomError
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	// 1. 检查是否是 sql.ErrNoRows（记录不存在）
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoRows
	}

	// 2. 检查是否是 PostgreSQL 特定错误
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return parsePgError(pqErr)
	}

	// 3. 其他数据库错误原样返回
	return err
}

// parsePgError 解析 PostgreSQL 特定错误
func parsePgError(pqErr *pq.Error) error {
	switch pqErr.Code {
	case PgErrUniqueViolation:
		// 唯一键冲突 - 根据约束名称返回具体错误
		if pqErr.Constraint != "" {
			switch pqErr.Constraint {
			case "users_username_key":
				return ErrUsernameDuplicate
			case "users_email_key":
				return ErrEmailDuplicate
			case "pastes_short_link_key":
				return ErrShortLinkDuplicate
			}
		}
		return ErrDuplicateKey

	case PgErrForeignKeyViolation:
		return ErrForeignKey

	case PgErrCheckViolation:
		return ErrCheckViolation

	case PgErrNotNullViolation:
		return ErrNotNullViolation

	case PgErrInvalidTextRepresentation, PgErrNumericValueOutOfRange:
		return ErrInvalidData

	case PgErrConnectionException, PgErrConnectionDoesNotExist, PgErrConnectionFailure:
		return ErrConnection

	case PgErrUndefinedColumn, PgErrUndefinedTable, PgErrSyntaxError:
		// 开发错误，返回原始错误
		return pqErr

	default:
		// 未知的 PostgreSQL 错误
		return pqErr
	}
}

// IsDuplicateKeyError 检查是否是唯一键冲突错误
func IsDuplicateKeyError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == PgErrUniqueViolation
	}
	return false
}

// IsForeignKeyError 检查是否是外键约束错误
func IsForeignKeyError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == PgErrForeignKeyViolation
	}
	return false
}

// IsNotFoundError 检查是否是记录不存在错误
func IsNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// IsConnectionError 检查是否是连接错误
func IsConnectionError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		code := string(pqErr.Code)
		return code == PgErrConnectionException ||
			code == PgErrConnectionDoesNotExist ||
			code == PgErrConnectionFailure
	}
	return false
}

// WrapDBError 包装数据库错误，保留原始错误信息
// 适用于需要在错误链中添加上下文的场景
func WrapDBError(err error, operation string) error {
	if err == nil {
		return nil
	}
	return errors.New(operation + ": " + err.Error())
}
