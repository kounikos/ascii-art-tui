package renderer

import (
	"strings"
	"testing"
)

func makeFont(chars string, lines [8]string) map[rune][]string {
	font := map[rune][]string{}
	for _, ch := range chars {
		font[ch] = lines[:]
	}
	return font
}

func TestRenderText_Empty(t *testing.T) {
	result, err := RenderText("", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderText(t *testing.T) {
	// Mock ASCII map with a simple character 'A'
	mockFont := map[rune][]string{
		'A': {
			"  A  ",
			" A A ",
			"AAAAA",
			"A   A",
			"A   A",
			"     ",
			"     ",
			"     ",
		},
	}

	cfg := &ColorConfig{
		Rules: []ColorRule{
			{ColorHex: "#FF0000", Substring: "A"},
		},
	}

	t.Run("Valid Text Simple", func(t *testing.T) {
		output, err := RenderText("A", mockFont, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if output == "" {
			t.Fatal("expected non-empty output")
		}
	})

	t.Run("With Color Config", func(t *testing.T) {
		output, err := RenderText("A", mockFont, cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		
		// The ASCII representation should now contain ansi markup!
		if output == "" {
			t.Fatal("expected non-empty output")
		}
	})

	t.Run("Empty Text", func(t *testing.T) {
		output, err := RenderText("", mockFont, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if output != "" {
			t.Fatalf("expected empty output, got %q", output)
		}
	})
}

func TestRenderText_Newline(t *testing.T) {
	font := makeFont("a", [8]string{"1", "2", "3", "4", "5", "6", "7", "8"})
	result, err := RenderText("a\\nb", font, nil)
	if err == nil && !strings.Contains(result, "\n") {
		t.Error("expected newline in result")
	}
}

func TestRenderText_UnsupportedChar(t *testing.T) {
	font := map[rune][]string{}
	_, err := RenderText("a", font, nil)
	if err == nil {
		t.Error("expected error for unsupported character")
	}
}

func TestRenderText_OnlyNewlines(t *testing.T) {
	result, err := RenderText("\\n\\n", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "\n\n" {
		t.Errorf("expected two newlines, got %q", result)
	}
}
