package log

import (
	"testing"
)

func Test_Debug(t *testing.T) {
	Debug("测试Debug")
}

func Test_Info(t *testing.T) {
	Info("测试Info")
}

func Test_Warn(t *testing.T) {
	Warn("测试Warn")
}

func Test_Error(t *testing.T) {
	Error("测试Error")
}
