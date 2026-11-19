# ✨ ASCII Studio TUI

![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat-square&logo=go)
![Bubbletea Version](https://img.shields.io/badge/BubbleTea-Layout-FF7CCB?style=flat-square)
![Lipgloss Themes](https://img.shields.io/badge/Lipgloss-Themed-B4BEFE?style=flat-square)

**ASCII Studio TUI** is a fully standalone, premium Terminal User Interface (TUI) built around a powerful ASCII art rendering engine. Developed using Go's `charmbracelet/bubbletea` ecosystem, it is designed from the ground up to provide a native, desktop-like editing experience directly within your CLI.

## ✨ Features

- **Dynamic Output Resizing**: The application automatically tracks your terminal dimensions. It restructures its grid layout in real-time—transitioning smoothly from a Side-By-Side layout on wider displays to a stacked vertical layout on smaller shells—preventing your ASCII art from clipping.
- **Highlighting Rules Manager**: Built like a professional text filter, the TUI allows you to construct dynamic arrays of active matching rules. You can define multiple substrings and associate them with independent CSS color formats, seamlessly applying overlapping styles across complex block characters.
- **Application Flow States**: The project defines intelligent application lifecycles mimicking robust desktop interactions, completely managing boot sequences (with a loading spinner) and graceful shutdown fade-out animations.
- **Vivid Intuitive Styling**: Enforces a vibrant Color Palette with glow states and border indicators. The interface securely pulls user focus across active editing panels to make dashboard navigation completely effortless.

---

## 🚀 Quick Start

Ensure you have a recent version of Go installed. Standard ANSI output support in your terminal emulator is required to view colors correctly.

```bash
# Clone the repository
git clone https://github.com/YongPadawan/ascii-art-tui
cd ascii-art-tui

# Boot the Dashboard
go run main.go
```

## ⌨️ TUI Navigation

The interface utilizes standard modern keybindings to ensure rapid interaction without touching your mouse.

| Action | Hotkey | Description |
| :--- | :---: | :--- |
| **Shift Focus** | `<TAB>` | Cycle through Editor inputs sequentially, or use `<SHIFT+TAB>` to traverse backwards. |
| **Add Rule** | `<CTRL+A>` | Spawn an additional matching Rule directly underneath your active highlight list. |
| **Drop Rule** | `<CTRL+X>` | Delete the last instantiated matching Rule from the ruleset. |
| **Change Theme** | `<CTRL+T>` | Instantly switch your ASCII generation format between Standard, Shadow, and Thinkertoy. |
| **Shutdown** | `<ESC>`/`<CTRL+C>` | Request a graceful exit sequence. |

---

## 🛠 Project Architecture

This application cleanly separates UI components from core processing algorithms.

### `main.go`
Controls rendering bounds and handles direct user interaction bindings through the `bubbletea` event loop.

### `internal/`
*Core dependencies tracking validation and formatting, entirely decoupled from the BubbleTea dependencies:*
- **`renderer/`**: Responsible for manipulating `ColorRule` slices and calculating overlapping CSS tag precedence during the ASCII assembly phase.
- **`ascii/`**: Houses bounds validation constraints to securely scrub input schemas and triggers banner execution.
- **`parser/`**: A fast, standard scanner that breaks `.txt` visual typography source files into localized memory arrays.
