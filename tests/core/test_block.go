package test_core

import (
	"testing"

	"github.com/PiperFinance/BS/src/core/events"
)

func TestParseLogs(t *testing.T) {

}

func TestEventTopicHash(t testing.T) {
	var tests = []struct {
		_case, want string
	}{
		{"asd", "asd"},
		{"asd", "asd"},
	}
	for _, tt := range tests {
		if ans := events.EventTopicHash(tt._case); tt.want != ans {
			t.Errorf("Expected [%s] %s, got %s", tt._case, tt.want, ans)
		}
	}
}
