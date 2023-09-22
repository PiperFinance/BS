package test_core

import (
	"testing"

	"github.com/PiperFinance/BS/src/core/utils"
)

func TestParseLogs(t *testing.T) {
}

func TestEventTopicHash(t testing.T) {
	tests := []struct {
		_case, want string
	}{
		{"asd", "asd"},
		{"asd", "asd"},
	}
	for _, tt := range tests {
		if ans := utils.EventTopicSignature(tt._case); tt.want != ans {
			t.Errorf("Expected [%s] %s, got %s", tt._case, tt.want, ans)
		}
	}
}
