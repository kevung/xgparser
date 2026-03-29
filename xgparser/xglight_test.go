package xgparser

import (
	"testing"
)

func TestStripRTF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text passthrough",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "simple RTF",
			input:    "{\\rtf1\\ansi\\ansicpg1252\\uc1\\deff0\\deflang1033\\deflangfe1033{\\fonttbl{\\f0\\fcharset0 Arial;}}\r\n{\\colortbl;}\r\n{\\*\\generator Wine Riched20 2.0;}\r\n\\pard\\sl-240\\slmult1 \\lang1033\\fs20\\f0 coup assez difficile\\par\r\n}",
			expected: "coup assez difficile",
		},
		{
			name:     "RTF with hex escape",
			input:    "{\\rtf1\\ansi\\ansicpg1252\\uc1\\deff0\\deflang1033\\deflangfe1033{\\fonttbl{\\f0\\fcharset0 Arial;}}\r\n{\\colortbl;}\r\n{\\*\\generator Wine Riched20 2.0;}\r\n\\pard\\sl-240\\slmult1 \\lang1033\\fs20\\f0 les double 66 sont toujpurs difficiles \\'e0 trouver\\par\r\n}",
			expected: "les double 66 sont toujpurs difficiles à trouver",
		},
		{
			name:     "RTF with escaped braces",
			input:    `{\rtf1 hello \{ world \}}`,
			expected: "hello { world }",
		},
		{
			name:     "RTF with escaped backslash",
			input:    `{\rtf1 hello \\ world}`,
			expected: `hello \ world`,
		},
		{
			name:     "RTF with multiple paragraphs",
			input:    `{\rtf1 first paragraph\par second paragraph\par}`,
			expected: "first paragraph\nsecond paragraph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripRTF(tt.input)
			if got != tt.expected {
				t.Errorf("stripRTF() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseCommentSegment(t *testing.T) {
	// Simulate a comment segment with two RTF comments separated by CRLF
	segment := `{\rtf1\ansi\ansicpg1252\uc1\deff0\deflang1033\deflangfe1033{\fonttbl{\f0\fcharset0 Arial;}}` + "\x01\x02" +
		`{\colortbl;}` + "\x01\x02" +
		`{\*\generator Wine Riched20 2.0;}` + "\x01\x02" +
		`\pard\sl-240\slmult1 \lang1033\fs20\f0 coup assez difficile\par` + "\x01\x02" +
		`}` + "\r\n" +
		`{\rtf1\ansi\ansicpg1252\uc1\deff0\deflang1033\deflangfe1033{\fonttbl{\f0\fcharset0 Arial;}}` + "\x01\x02" +
		`{\colortbl;}` + "\x01\x02" +
		`{\*\generator Wine Riched20 2.0;}` + "\x01\x02" +
		`\pard\sl-240\slmult1 \lang1033\fs20\f0 les double 66 sont toujpurs difficiles \'e0 trouver\par` + "\x01\x02" +
		`}` + "\r\n"

	comments := parseCommentSegment([]byte(segment))

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	expected := []string{
		"coup assez difficile",
		"les double 66 sont toujpurs difficiles à trouver",
	}

	for i, exp := range expected {
		if comments[i] != exp {
			t.Errorf("comment[%d] = %q, want %q", i, comments[i], exp)
		}
	}
}
