package valueobject

import (
	"errors"
	"net/mail"
)

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
)

// Email 邮箱值对象
type Email struct {
	value string
}

// NewEmail 创建邮箱值对象
func NewEmail(email string) (*Email, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, ErrInvalidEmailFormat
	}
	return &Email{value: email}, nil
}

// Value 获取邮箱值
func (e Email) Value() string {
	return e.value
}

// String 字符串表示
func (e Email) String() string {
	return e.value
}

// Equals 比较两个邮箱是否相等
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
