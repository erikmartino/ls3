package main

import (
	"fmt"

	"github.com/rivo/tview"
)

// showProgressWindow displays a progress window for file download
func showProgressWindow(app *tview.Application, filename string, onCancel func()) (*tview.Modal, func(current, total int64)) {
	cancelled := false

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Downloading: %s\n\nPreparing download...", filename)).
		AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Cancel" && onCancel != nil {
				cancelled = true
				onCancel()
			}
		})

	updateProgress := func(current, total int64) {
		if cancelled {
			return
		}

		app.QueueUpdateDraw(func() {
			var progressText string

			if total > 0 {
				percentage := float64(current) * 100.0 / float64(total)
				barWidth := 40
				filled := int(percentage * float64(barWidth) / 100.0)

				bar := "["
				for i := 0; i < barWidth; i++ {
					if i < filled {
						bar += "█"
					} else {
						bar += "░"
					}
				}
				bar += "]"

				progressText = fmt.Sprintf("Downloading: %s\n\n%s\n%.1f%% (%s / %s)",
					filename,
					bar,
					percentage,
					formatBytes(current),
					formatBytes(total))
			} else {
				bar := "[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░]"
				progressText = fmt.Sprintf("Downloading: %s\n\n%s\nDownloading... (%s)",
					filename,
					bar,
					formatBytes(current))
			}

			modal.SetText(progressText)
		})
	}

	return modal, updateProgress
}
