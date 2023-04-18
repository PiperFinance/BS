package test_tasks

import (
	"testing"

	"github.com/PiperFinance/BS/src/core/tasks"
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
		if ans := tasks.EventTopicHash(tt._case); tt.want != ans {
			t.Errorf("Expected [%s] %s, got %s", tt._case, tt.want, ans)
		}
	}
}
