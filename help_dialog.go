package main

import (
	"fmt"

	"github.com/rivo/tview"
)

// showHelpDialog displays a help dialog with all available shortcuts
func showHelpDialog(app *tview.Application) *tview.Modal {
	// Use fixed-width formatting for better alignment
	helpText := fmt.Sprintf(`[yellow]ls3 - S3 Browser Shortcuts[-]

[cyan]Navigation:[-]
  %-15s %s
  %-15s %s
  %-15s %s
  %-15s %s

[cyan]File Operations:[-]
  %-15s %s
  %-15s %s
  %-15s %s

[cyan]Application:[-]
  %-15s %s
  %-15s %s
  %-15s %s

[cyan]File Viewing:[-]
  %-15s %s

[cyan]Features:[-]
  • ASCII art preview for images
  • Gzip decompression for compressed files
  • Progress window for downloads with cancel option
  • Session state persistence
  • Command line S3 URL support

Press ESC or Enter to close this help.`,
		"[white]↑/↓[-]", "Navigate up/down in lists",
		"[white]←/Backspace[-]", "Go back / up one level",
		"[white]→/Enter[-]", "Enter directory / view file",
		"[white]Ctrl+L[-]", "Refresh current view",
		"[white]c[-]", "Copy S3 URL to clipboard",
		"[white]C[-]", "Copy presigned URL to clipboard",
		"[white]d[-]", "Download file to current directory",
		"[white]?[-]", "Show this help dialog",
		"[white]Ctrl+C[-]", "Exit application (prints current S3 URL)",
		"[white]ESC[-]", "Close dialogs / go back",
		"[white]ESC/←[-]", "Return to file browser from file view")

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// The modal will be removed by the caller
		})

	return modal
}
