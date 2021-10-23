package launcher

import "testing"

func Test_stripEscapeSeqs(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"Begins with 1b", "\u001b[55mABC", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripEscapeSeqs(tt.args); got != tt.want {
				t.Errorf("stripEscapeSeqs() = %q, want %q", got, tt.want)
			}
		})
	}
}
