// Package renderer is responsible for assembling textual characters into their multi-line ASCII representations and applying standard or layered Lipgloss coloring organically across the map.
package renderer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorRule defines a singular constraint mapping a raw unstyled Substring explicitly into a designated CSS ColorHex.
type ColorRule struct {
	Substring string
	ColorHex  string
}

// ColorConfig structurally holds a dynamic slice of overlapping ColorRules applied explicitly natively through Top-Bottom precedence within the renderer.
type ColorConfig struct {
	Rules []ColorRule
}

// RenderText takes the input text, a map of ASCII characters, and an optional color configuration
func RenderText(text string, asciiMap map[rune][]string, cfg *ColorConfig) (string, error) {
	if text == "" {
		return "", nil
	}

	subtext := strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(subtext, "\\n")

	allEmpty := true
	for _, part := range parts {
		if part != "" {
			allEmpty = false
			break
		}
	}

	var result strings.Builder

	if allEmpty {
		for i := 0; i < len(parts)-1; i++ {
			result.WriteString("\n")
		}
		return result.String(), nil
	}

	for _, part := range parts {
		if part == "" {
			result.WriteString("\n")
			continue
		}

		lineRunes := []rune(part)
		charStyles := make([]lipgloss.Style, len(lineRunes))

		// Process layering rules if they exist
		if cfg != nil && len(cfg.Rules) > 0 {
			for _, rule := range cfg.Rules {
				if rule.ColorHex == "" {
					continue
				}

				style := lipgloss.NewStyle().Foreground(lipgloss.Color(rule.ColorHex))

				if rule.Substring == "" {
					// Base rule coloring everything
					for i := range charStyles {
						charStyles[i] = style
					}
				} else {
					subRunes := []rune(rule.Substring)
					subLen := len(subRunes)
					if subLen == 0 {
						continue
					}

					for j := 0; j <= len(lineRunes)-subLen; j++ {
						match := true
						for k := 0; k < subLen; k++ {
							if lineRunes[j+k] != subRunes[k] {
								match = false
								break
							}
						}
						// Apply colors, inherently overriding any previous rules because of loop precedence!
						if match {
							for k := 0; k < subLen; k++ {
								charStyles[j+k] = style
							}
						}
					}
				}
			}
		}

		for i := 0; i < 8; i++ {
			for charIndex, r := range lineRunes {
				lines, ok := asciiMap[r]
				if !ok {
					return "", fmt.Errorf("character %q not found in map", r)
				}

				segment := lines[i]

				// If a style is defined (not empty), apply it securely!
				if charStyles[charIndex].GetForeground() != lipgloss.Color("") {
					segment = charStyles[charIndex].Render(segment)
				}
				result.WriteString(segment)
			}
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}
