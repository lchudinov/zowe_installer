package launcher

import "testing"

func Test_strip(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"Begins with 1b", "\u001b[55mABC", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := strip(tt.args); got != tt.want {
				t.Errorf("strip() = %q, want %q", got, tt.want)
			}
		})
	}
}
