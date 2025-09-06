package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Fetch S3 buckets
	result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("failed to list buckets: %v", err)
	}

	// Create TUI application
	app := tview.NewApplication()
	list := tview.NewList()
	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Select an S3 bucket")

	// Populate the list with S3 buckets
	for _, bucket := range result.Buckets {
		bucketName := *bucket.Name
		list.AddItem(bucketName, "", 0, func() {
			text.SetText(fmt.Sprintf("s3://%s", bucketName))
		})
	}

	// Add a "Quit" option
	list.AddItem("Quit", "", 'q', func() {
		app.Stop()
	})

	// Hide secondary text to remove blank lines
	list.ShowSecondaryText(false)

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 3, 1, false).
		AddItem(list, 0, 1, true)

	// Keybindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
		}
		return event
	})

	// Run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
