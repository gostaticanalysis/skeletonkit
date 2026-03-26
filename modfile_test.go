package skeletonkit

import "testing"

func Test_parseGoVersion(t *testing.T) {
	tests := []struct {
		v    string
		want string
	}{
		{"", ""},
		{"go", ""},
		{"go1", "1"},
		{"go1.20", "1.20"},
		{"go1.21.0", "1.21.0"},
		{"go1.23rc1", "1.23rc1"},
		{"devel go1.24-f99f5da18f Thu Nov 14 22:29:26 2024 +0000 darwin/arm64", "1.24"},
	}
	for _, tt := range tests {
		t.Run(tt.v, func(t *testing.T) {
			version := parseGoVersion(tt.v)
			if version != tt.want {
				t.Errorf("parseGoVersion() = %v, want %v", version, tt.want)
			}
		})
	}
}
