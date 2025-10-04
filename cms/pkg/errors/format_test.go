package errors

import "testing"

func TestRemoveLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		target string
		prefix string
		want   string
	}{
		{"/eins/zwei", "/eins", "/zwei"},
		{"/eins/zwei", "/eins/", "zwei"},
		{"/", "/eins", "/"},
		{"/eins/zwei/", "/eins/zwei/", "/eins/zwei/"},
		{"/eins/zwei/", "/eins/drei/", "zwei/"},
		{"/eins/zwei/", "./", "/eins/zwei/"},
	}
	for _, tt := range tests {
		got := removeLongestCommonPrefix(tt.target, tt.prefix)
		if got != tt.want {
			t.Errorf("SplitPath() gotHead = %v, want %v", got, tt.want)
		}
	}
}
