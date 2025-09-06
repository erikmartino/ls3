package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AppState holds the current state of the application
type AppState struct {
	CurrentBucket string `json:"current_bucket"`
	CurrentPrefix string `json:"current_prefix"`
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ls3_state.json"), nil
}

// saveState saves the current application state to a config file
func saveState(state AppState) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// loadState loads the application state from the config file
func loadState() (AppState, error) {
	var state AppState
	configPath, err := getConfigPath()
	if err != nil {
		return state, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return empty state
			return state, nil
		}
		return state, err
	}

	err = json.Unmarshal(data, &state)
	return state, err
}

func main() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	// Load saved state
	savedState, err := loadState()
	if err != nil {
		log.Printf("failed to load state: %v", err)
	}

	// Track current state
	currentState := savedState

	// Create TUI application
	app := tview.NewApplication()
	list := tview.NewList()
	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Select an S3 bucket")
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if mainText != "Quit" {
			text.SetText(fmt.Sprintf("s3://%s", mainText))
		} else {
			text.SetText("Select an S3 bucket")
		}
	})

	// Fetch S3 buckets and populate the list
	go func() {
		buckets, err := getBuckets(context.TODO(), client)
		if err != nil {
			log.Fatalf("failed to list buckets: %v", err)
		}

		app.QueueUpdateDraw(func() {
			for _, bucket := range buckets {
				bucketName := *bucket.Name
				list.AddItem(bucketName, "", 0, func() {
					text.SetText(fmt.Sprintf("s3://%s", bucketName))
				})
			}

			// We'll handle saved state navigation after listObjects is defined
		})
	}()

	// Hide secondary text to remove blank lines
	list.ShowSecondaryText(false)

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 3, 1, false).
		AddItem(list, 0, 1, true)

	var showFileContent func(bucketName, objectKey string, previousFlex *tview.Flex)

	// Function to list objects in a bucket
	var listObjects func(bucketName, prefix string)
	showFileContent = func(bucketName, objectKey string, previousFlex *tview.Flex) {
		// Update current state
		currentState.CurrentBucket = bucketName
		currentState.CurrentPrefix = strings.TrimSuffix(objectKey, filepath.Base(objectKey))
		saveState(currentState)
		textView := tview.NewTextView().
			SetText("Loading file content...").
			SetDynamicColors(true)

		textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyLeft {
				app.SetRoot(previousFlex, true)
				return nil
			}
			return event
		})

		go func() {
			body, err := getObjectContent(context.TODO(), client, bucketName, objectKey)
			if err != nil {
				app.QueueUpdateDraw(func() {
					textView.SetText(fmt.Sprintf("Error: %v", err))
				})
				return
			}

			app.QueueUpdateDraw(func() {
				textView.SetText(string(body))
			})
		}()

		app.SetRoot(textView, true)
	}
	listObjects = func(bucketName, prefix string) {
		// Update current state
		currentState.CurrentBucket = bucketName
		currentState.CurrentPrefix = prefix
		saveState(currentState)
		currentPath := fmt.Sprintf("s3://%s/%s", bucketName, prefix)
		text.SetText(currentPath)
		objectList := tview.NewList()
		objectList.ShowSecondaryText(false)
		objectList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
			path := fmt.Sprintf("s3://%s/%s", bucketName, mainText)
			text.SetText(path)
		})

		objectFlex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 3, 1, false).
			AddItem(objectList, 0, 1, true)

		objectList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyLeft {
				if prefix != "" {
					newPrefix := prefix[:len(prefix)-1]
					lastSlash := strings.LastIndex(newPrefix, "/")
					if lastSlash != -1 {
						listObjects(bucketName, newPrefix[:lastSlash+1])
					} else {
						listObjects(bucketName, "")
					}
				} else {
					app.SetRoot(flex, true)
				}
				return nil
			} else if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyRight {
				selectedItem := objectList.GetCurrentItem()
				if selectedItem >= 0 {
					itemName, _ := objectList.GetItemText(selectedItem)
					if strings.HasSuffix(itemName, "/") {
						listObjects(bucketName, itemName)
					} else {
						showFileContent(bucketName, itemName, objectFlex)
					}
				}
				return nil
			}
			return event
		})

		app.SetRoot(objectFlex, true)

		go func() {
			objects, err := listS3Objects(context.TODO(), client, bucketName, prefix)
			if err != nil {
				log.Printf("failed to list objects: %v", err)
				return
			}

			app.QueueUpdateDraw(func() {
				for _, p := range objects.CommonPrefixes {
					objectList.AddItem(*p.Prefix, "", 0, nil)
				}
				for _, o := range objects.Contents {
					if *o.Key != prefix {
						objectList.AddItem(*o.Key, "", 0, nil)
					}
				}
			})
		}()
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyRight {
			selectedItem := list.GetCurrentItem()
			if selectedItem >= 0 && selectedItem < list.GetItemCount()-1 {
				bucketName, _ := list.GetItemText(selectedItem)
				listObjects(bucketName, "")
			}
		}
		return event
	})

	// Keybindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			saveState(currentState)
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			saveState(currentState)
			app.Stop()
			return nil
		}
		return event
	})

	// Handle saved state navigation after everything is set up
	if savedState.CurrentBucket != "" {
		go func() {
			// Wait for buckets to be loaded first
			buckets, err := getBuckets(context.TODO(), client)
			if err != nil {
				return
			}

			// Check if the saved bucket still exists
			bucketExists := false
			for _, bucket := range buckets {
				if *bucket.Name == savedState.CurrentBucket {
					bucketExists = true
					break
				}
			}

			if bucketExists {
				app.QueueUpdateDraw(func() {
					listObjects(savedState.CurrentBucket, savedState.CurrentPrefix)
				})
			}
		}()
	}

	// Run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
