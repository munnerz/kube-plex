// ffmpeg package contains helpers to unescape and escape FFmpeg strings
package ffmpeg

import (
	"testing"
)

func TestUnescape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"string that needs no escaping", "lorem ipsum", "lorem ipsum"},
		{"escaped with backslash", "lorem\\ ipsum", "lorem ipsum"},
		{"escaped with quotes", "'lorem ipsum'", "lorem ipsum"},
		{"backslash within quotes", "'lorem\\ ipsum'", "lorem\\ ipsum"},
		{"different escapes in same", "lorem\\ ipsum 'lorem ipsum'", "lorem ipsum lorem ipsum"},
		{"escaped quotes", "lorem\\' ipsum\\'", "lorem' ipsum'"},
		{"mixed escapes", "lorem\\' ip'su\\'m", "lorem' ipsu\\m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unescape(tt.in); got != tt.want {
				t.Errorf("Unescape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"escapes backslash", "lorem\\ipsum", "lorem\\\\ipsum"},
		{"escapes quotation marks", "lorem'ipsum", "lorem\\'ipsum"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Escape(tt.in); got != tt.want {
				t.Errorf("Escape() = %v, want %v", got, tt.want)
			}
		})
	}
}
