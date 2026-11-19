// Package ascii exposes the core text-to-ASCII generation engine, ensuring validation bounds natively.
package ascii

import (
    "fmt"
    "strings"

    "ascii-art-tui/internal/parser"
    "ascii-art-tui/internal/renderer"
)

// validBanners contains the supported ASCII banner themes.
var validBanners = map[string]bool{"standard": true, "shadow": true, "thinkertoy": true}

// GenerateAscii transforms an input string into a multi-line ASCII art string.
// It rigorously validates character bounds, resolves explicit literal linebreaks ("\n"),
// strictly requests underlying layout parsing, and proxies styles down to the rendering interface.
func GenerateAscii(text string, banner string, cfg *renderer.ColorConfig) (string, error) {
    if text == "" {
        return "", nil
    }

    if !validBanners[banner] {
        return "", fmt.Errorf("invalid banner: %q", banner)
    }

    text = strings.ReplaceAll(text, "\r\n", "\n")
    text = strings.ReplaceAll(text, "\\n", "\n") // Support literal '\n' characters as line breaks

    for _, ch := range text {
        if (ch < ' ' || ch > '~') && ch != '\n' {
            return "", fmt.Errorf("unsupported character: %q", ch)
        }
    }

    font, err := parser.ParseBanner("ascii-style/" + banner + ".txt")
    if err != nil {
        return "", fmt.Errorf("banner not found: %w", err)
    }

    return renderer.RenderText(text, font, cfg)
}
