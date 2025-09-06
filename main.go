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

	// Create TUI application
	app := tview.NewApplication()
	list := tview.NewList()
	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Select an S3 bucket")

	// Fetch S3 buckets and populate the list
	go func() {
		result, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
		if err != nil {
			log.Fatalf("failed to list buckets: %v", err)
		}

		app.QueueUpdateDraw(func() {
			for _, bucket := range result.Buckets {
				bucketName := *bucket.Name
				list.AddItem(bucketName, "", 0, func() {
					text.SetText(fmt.Sprintf("s3://%s", bucketName))
				})
			}
			list.AddItem("Quit", "", 'q', func() {
				app.Stop()
			})
		})
	}()

	// Hide secondary text to remove blank lines
	list.ShowSecondaryText(false)

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 3, 1, false).
		AddItem(list, 0, 1, true)

	// Function to list objects in a bucket
	listObjects := func(bucketName string) {
		text.SetText(fmt.Sprintf("s3://%s", bucketName))
		objectList := tview.NewList()

		objectList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyLeft {
				app.SetRoot(flex, true)
			}
			return event
		})

		objectFlex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 3, 1, false).
			AddItem(objectList, 0, 1, true)

		app.SetRoot(objectFlex, true)

		// Fetch objects from the selected bucket
		go func() {
			objects, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
				Bucket: &bucketName,
			})
			if err != nil {
				log.Printf("failed to list objects: %v", err)
				return
			}

			app.QueueUpdateDraw(func() {
				for _, object := range objects.Contents {
					objectList.AddItem(*object.Key, "", 0, nil)
				}
			})
		}()
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyRight {
			selectedItem := list.GetCurrentItem()
			if selectedItem >= 0 && selectedItem < list.GetItemCount()-1 {
				bucketName, _ := list.GetItemText(selectedItem)
				listObjects(bucketName)
			}
		}
		return event
	})

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
