package valueobject

import (
	"github.com/falconfan123/Go-mall/common/utils/cryptx"
)

// PasswordHash 密码哈希值对象
type PasswordHash struct {
	value string
}

// NewPasswordHash 从明文密码创建密码哈希
func NewPasswordHash(plainPassword string) *PasswordHash {
	hashed := cryptx.PasswordEncrypt(plainPassword)
	return &PasswordHash{value: hashed}
}

// NewPasswordHashFromHash 从已有哈希创建密码哈希对象
func NewPasswordHashFromHash(hash string) *PasswordHash {
	return &PasswordHash{value: hash}
}

// Value 获取哈希值
func (p PasswordHash) Value() string {
	return p.value
}

// Verify 验证密码是否正确
func (p PasswordHash) Verify(plainPassword string) bool {
	return cryptx.PasswordVerify(plainPassword, p.value)
}

// Equals 比较两个哈希是否相等
func (p PasswordHash) Equals(other PasswordHash) bool {
	return p.value == other.value
}
