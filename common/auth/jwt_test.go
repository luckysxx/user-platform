package auth

// ============================================================
// Go 测试入门指南 - 以 JWT 为例
// ============================================================
//
// 【核心规则】
// 1. 测试文件必须以 _test.go 结尾，Go 编译器会自动识别
// 2. 测试文件和被测文件放在同一个包（同一目录）下
// 3. 测试函数必须以 Test 开头，接收 *testing.T 参数
// 4. 运行命令: go test ./...（跑所有测试）或 go test ./common/auth/（跑这个包）
//
// 【文件命名】
//   被测文件:  jwt.go
//   测试文件:  jwt_test.go
//
// 【断言方式】
//   Go 标准库没有 assert，用 if + t.Errorf / t.Fatalf 来判断
//   t.Errorf  -> 报告错误但继续执行后续测试
//   t.Fatalf  -> 报告错误并立即终止当前测试函数
//
// 【表驱动测试 Table-Driven Test】
//   Go 社区最推荐的写法，把多组测试用例放进一个切片，
//   用 for + t.Run 遍历执行，清晰且易于扩展
// ============================================================

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --------------------------------------------------
// 测试 1: 最基本的测试 - 能正常生成和解析 Token
// --------------------------------------------------
// 函数命名: Test + 被测函数名 + 场景描述
func TestGenerateAndParseToken_Success(t *testing.T) {
	// -------- Arrange(准备) --------
	// 创建一个 JWTManager，传一个测试用的密钥
	manager := NewJWTManager("test-secret-key")
	userID := int64(42)

	// -------- Act(执行) --------
	// 生成 Token
	tokenString, err := manager.GenerateToken(userID)

	// -------- Assert(断言) --------
	// 1) 生成不应该报错
	if err != nil {
		// t.Fatalf: 遇到致命错误直接终止，因为后面的解析也没意义了
		t.Fatalf("GenerateToken() 报错了: %v", err)
	}

	// 2) Token 不应该是空字符串
	if tokenString == "" {
		t.Fatal("GenerateToken() 返回了空字符串")
	}

	// 3) 用同一个 manager 解析 Token
	claims, err := manager.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("ParseToken() 报错了: %v", err)
	}

	// 4) 解析出来的 UserID 应该和生成时一致
	if claims.UserID != userID {
		// t.Errorf: 报告错误但继续跑，用于非致命的断言
		t.Errorf("UserID 不对, 期望 %d, 实际 %d", userID, claims.UserID)
	}

	// 5) Issuer 应该是 "gopher-paste"
	if claims.Issuer != "gopher-paste" {
		t.Errorf("Issuer 不对, 期望 %q, 实际 %q", "gopher-paste", claims.Issuer)
	}
}

// --------------------------------------------------
// 测试 2: 用错误的密钥解析 Token，应该失败
// --------------------------------------------------
func TestParseToken_WrongSecret(t *testing.T) {
	// 用密钥 A 生成
	managerA := NewJWTManager("secret-A")
	tokenString, err := managerA.GenerateToken(1)
	if err != nil {
		t.Fatalf("生成 Token 失败: %v", err)
	}

	// 用密钥 B 解析 -> 应该报错
	managerB := NewJWTManager("secret-B")
	_, err = managerB.ParseToken(tokenString)

	// 我们期望 err 不为 nil（解析应该失败）
	if err == nil {
		t.Error("用错误密钥解析应该失败，但没有报错")
	}
}

// --------------------------------------------------
// 测试 3: 解析一个乱写的字符串，应该失败
// --------------------------------------------------
func TestParseToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")
	_, err := manager.ParseToken("not-a-valid-token")

	if err == nil {
		t.Error("解析非法字符串应该失败，但没有报错")
	}
}

// --------------------------------------------------
// 测试 4: 表驱动测试 (Table-Driven Test)
// 这是 Go 社区最推荐的模式，一次测试多种情况
// --------------------------------------------------
func TestParseToken_TableDriven(t *testing.T) {
	manager := NewJWTManager("my-secret")

	// 定义测试用例表，每一行就是一个测试场景
	tests := []struct {
		name        string // 测试名称（会显示在输出里，方便定位失败）
		token       string // 输入
		expectError bool   // 是否期望报错
	}{
		{
			name:        "空字符串",
			token:       "",
			expectError: true,
		},
		{
			name:        "随机字符串",
			token:       "abc123",
			expectError: true,
		},
		{
			name:        "格式像JWT但内容错误",
			token:       "eyJ.eyJ.sig",
			expectError: true,
		},
	}

	// 遍历每个用例，用 t.Run 创建子测试
	for _, tt := range tests {
		// t.Run 的第一个参数是子测试名称
		// 运行时输出类似: TestParseToken_TableDriven/空字符串
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.ParseToken(tt.token)

			if tt.expectError && err == nil {
				t.Errorf("期望报错，但 err 是 nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("不期望报错，但 err = %v", err)
			}
		})
	}
}

// --------------------------------------------------
// 测试 5: 验证过期 Token 会被拒绝
// --------------------------------------------------
func TestParseToken_Expired(t *testing.T) {
	manager := NewJWTManager("test-secret")

	// 手动创建一个已过期的 Token（过期时间设为 1 小时前）
	claims := Claims{
		UserID: 99,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "gopher-paste",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("创建过期Token失败: %v", err)
	}

	// 解析过期 Token -> 应该报错
	_, err = manager.ParseToken(tokenString)
	if err == nil {
		t.Error("解析过期 Token 应该失败，但没有报错")
	}
}

// --------------------------------------------------
// 测试 6: 多个 UserID 的表驱动测试
// --------------------------------------------------
func TestGenerateToken_MultipleUserIDs(t *testing.T) {
	manager := NewJWTManager("test-secret")

	userIDs := []int64{1, 100, 999999, 0}

	for _, id := range userIDs {
		t.Run(fmt.Sprintf("userID_%d", id), func(t *testing.T) {
			tokenStr, err := manager.GenerateToken(id)
			if err != nil {
				t.Fatalf("生成失败: %v", err)
			}

			claims, err := manager.ParseToken(tokenStr)
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			if claims.UserID != id {
				t.Errorf("UserID 不匹配, 期望 %d, 实际 %d", id, claims.UserID)
			}
		})
	}
}
