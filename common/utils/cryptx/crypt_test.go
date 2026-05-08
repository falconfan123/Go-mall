package cryptx

import (
	"testing"
)

func TestPasswordVerify(t *testing.T) {
	password := "test"
	hash := PasswordEncrypt(password)
	t.Logf("hash: %s", hash)
	if PasswordVerify(password, hash) {
		t.Log("密码验证成功")
	} else {
		t.Error("密码验证失败")
	}
}