package main

import (
	"github.com/rivo/tview"
)

// showHelpDialog displays a help dialog with all available shortcuts
func showHelpDialog(app *tview.Application) *tview.Modal {
	helpText := `[yellow]ls3 - S3 Browser Shortcuts[white]

[yellow]Navigation:[white]
  [green]↑/↓[white]        Navigate up/down in lists
  [green]←/Backspace[white] Go back / up one level
  [green]→/Enter[white]     Enter directory / view file
  [green]Ctrl+L[white]      Refresh current view

[yellow]File Operations:[white]
  [green]c[white]           Copy S3 URL to clipboard
  [green]d[white]           Download file to current directory

[yellow]Application:[white]
  [green]?[white]           Show this help dialog
  [green]Ctrl+C[white]      Exit application (prints current S3 URL)
  [green]ESC[white]         Close dialogs / go back

[yellow]File Viewing:[white]
  [green]ESC/←[white]       Return to file browser from file view

[yellow]Features:[white]
  • ASCII art preview for images
  • Gzip decompression for compressed files
  • Progress window for downloads with cancel option
  • Session state persistence
  • Command line S3 URL support

Press ESC or Enter to close this help.`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// The modal will be removed by the caller
		})

	return modal
}
