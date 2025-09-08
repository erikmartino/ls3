package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AppState holds the current state of the application
type AppState struct {
	CurrentBucket string `json:"current_bucket"`
	CurrentPrefix string `json:"current_prefix"`
}

// ObjectEntry holds information about an S3 object for display
type ObjectEntry struct {
	Key          string
	IsDirectory  bool
	Size         int64
	LastModified *time.Time
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

// formatFileSize formats a file size in bytes to human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDate formats a time to a readable date string
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// formatFileEntry formats a filename with right-aligned size and date metadata
func formatFileEntry(name string, size int64, date *time.Time, width int) string {
	if width <= 0 {
		width = 80 // default width
	}

	sizeStr := formatFileSize(size)
	dateStr := formatDate(date)
	metadata := fmt.Sprintf("%s  %s", sizeStr, dateStr)

	// Calculate available space for the name
	availableSpace := width - len(metadata) - 2 // 2 for padding
	if availableSpace < len(name) {
		// If name is too long, truncate it
		if availableSpace > 3 {
			name = name[:availableSpace-3] + "..."
		}
	}

	// Create format string with right-aligned metadata
	return fmt.Sprintf("%-*s %s", availableSpace, name, metadata)
}

// getTerminalWidth returns the current terminal width
func getTerminalWidth() int {
	// Try to get width from tput command
	cmd := exec.Command("tput", "cols")
	output, err := cmd.Output()
	if err == nil {
		if width, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil && width > 0 {
			return width
		}
	}

	// Try environment variables
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if width, err := strconv.Atoi(cols); err == nil && width > 0 {
			return width
		}
	}

	// Default fallback
	return 80
}

// formatDirEntry formats a directory name with right-aligned DIR indicator
func formatDirEntry(name string, width int) string {
	if width <= 0 {
		width = 80 // default width
	}

	metadata := "DIR"
	availableSpace := width - len(metadata) - 2 // 2 for padding
	if availableSpace < len(name) {
		// If name is too long, truncate it
		if availableSpace > 3 {
			name = name[:availableSpace-3] + "..."
		}
	}

	return fmt.Sprintf("%-*s %s", availableSpace, name, metadata)
}

func parseS3URL(url string) (bucket, prefix string, err error) {
	if !strings.HasPrefix(url, "s3://") {
		return "", "", fmt.Errorf("URL must start with s3://")
	}

	path := strings.TrimPrefix(url, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid S3 URL: missing bucket name")
	}

	bucket = parts[0]
	if len(parts) > 1 {
		prefix = parts[1]
		if !strings.HasSuffix(prefix, "/") && prefix != "" {
			prefix += "/"
		}
	}

	return bucket, prefix, nil
}

// isGzipped checks if the content is gzipped by checking the magic number
func isGzipped(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// decompressIfGzipped decompresses the data if it's gzipped, otherwise returns it as-is
func decompressIfGzipped(data []byte, filename string) ([]byte, error) {
	// Check by file extension first
	isGzipFile := strings.HasSuffix(strings.ToLower(filename), ".gz") ||
		strings.HasSuffix(strings.ToLower(filename), ".gzip")

	// Also check by content magic number
	hasGzipMagic := isGzipped(data)

	if !isGzipFile && !hasGzipMagic {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		// If gzip decompression fails, return original data
		return data, nil
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		// If reading fails, return original data
		return data, nil
	}

	return decompressed, nil
}

func main() {
	// Parse command line arguments
	flag.Parse()

	var s3URL string
	if len(flag.Args()) > 0 {
		s3URL = flag.Args()[0]
	}

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

	// Override with URL argument if provided
	if s3URL != "" {
		bucket, prefix, err := parseS3URL(s3URL)
		if err != nil {
			log.Fatalf("invalid S3 URL: %v", err)
		}
		currentState.CurrentBucket = bucket
		currentState.CurrentPrefix = prefix
	}

	// Create TUI application
	app := tview.NewApplication()
	bucketTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	text := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Select an S3 bucket")

	// Store bucket entries for proper navigation
	var bucketEntries []types.Bucket

	// Global variable to store the current refresh function for resize handling
	var currentRefreshFunc func()

	// Fetch S3 buckets and populate the table
	go func() {
		buckets, err := getBuckets(context.TODO(), client)
		if err != nil {
			log.Fatalf("failed to list buckets: %v", err)
		}

		app.QueueUpdateDraw(func() {
			// Clear and set up table headers
			bucketTable.Clear()
			bucketTable.SetCell(0, 0, tview.NewTableCell("Bucket Name").SetTextColor(tcell.ColorYellow).SetSelectable(false))
			bucketTable.SetCell(0, 1, tview.NewTableCell("Created").SetTextColor(tcell.ColorYellow).SetSelectable(false))

			bucketEntries = buckets
			row := 1
			for _, bucket := range buckets {
				bucketName := *bucket.Name
				creationDate := ""
				if bucket.CreationDate != nil {
					creationDate = bucket.CreationDate.Format("2006-01-02 15:04")
				}

				bucketTable.SetCell(row, 0, tview.NewTableCell(bucketName))
				bucketTable.SetCell(row, 1, tview.NewTableCell(creationDate))
				row++
			}

			// Select first bucket if available
			if len(buckets) > 0 {
				bucketTable.Select(1, 0)
				text.SetText(fmt.Sprintf("s3://%s", *buckets[0].Name))
			}
		})
	}()

	// Update path display when bucket selection changes
	bucketTable.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(bucketEntries) { // Skip header row
			bucketName := *bucketEntries[row-1].Name
			text.SetText(fmt.Sprintf("s3://%s", bucketName))
		}
	})

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 3, 1, false).
		AddItem(bucketTable, 0, 1, true)

	var showFileContent func(bucketName, objectKey string, previousFlex *tview.Flex)

	// Function to list objects in a bucket
	var listObjects func(bucketName, prefix string)
	showFileContent = func(bucketName, objectKey string, previousFlex *tview.Flex) {
		// Update current state
		currentState.CurrentBucket = bucketName
		currentState.CurrentPrefix = strings.TrimSuffix(objectKey, filepath.Base(objectKey))
		saveState(currentState)

		// Determine if this might be an image file for better loading message
		loadingMessage := "Loading file content..."
		if isImageFile(objectKey) {
			loadingMessage = "Loading image and converting to ASCII art..."
		}

		textView := tview.NewTextView().
			SetText(loadingMessage).
			SetDynamicColors(true)

		textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyLeft {
				app.SetRoot(previousFlex, true)
				return nil
			}
			if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
				// Scroll down a page
				row, col := textView.GetScrollOffset()
				_, _, _, height := textView.GetRect()
				textView.ScrollTo(row+height-1, col)
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

			// Decompress if gzipped
			decompressed, err := decompressIfGzipped(body, objectKey)
			if err != nil {
				app.QueueUpdateDraw(func() {
					textView.SetText(fmt.Sprintf("Error decompressing: %v", err))
				})
				return
			}

			app.QueueUpdateDraw(func() {
				// Get terminal dimensions for ASCII art
				_, _, width, height := textView.GetRect()
				if width == 0 {
					width = getTerminalWidth()
				}
				if height == 0 {
					height = 25 // reasonable default
				}

				// Try to convert to ASCII art if it's an image
				if ascii, isImage := convertToASCIIArt(decompressed, objectKey, width, height); isImage {
					textView.SetText("[green]ASCII Art Preview[white]\n\n" + ascii + "\n\n[yellow]Press ESC or Left Arrow to go back[white]")
				} else {
					// Display as regular text
					content := string(decompressed)
					if len(content) > 0 {
						textView.SetText(content)
					} else {
						textView.SetText("[yellow]File is empty or contains binary data[white]")
					}
				}
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
		// Create table with proper columns
		objectTable := tview.NewTable().
			SetBorders(false).
			SetSelectable(true, false)

		// Store object entries for proper key handling
		var objectEntries []ObjectEntry

		objectFlex := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(text, 3, 1, false).
			AddItem(objectTable, 0, 1, true)

		objectTable.SetSelectedFunc(func(row, column int) {
			if row > 0 && row-1 < len(objectEntries) { // Skip header row
				entry := objectEntries[row-1]
				if entry.IsDirectory {
					listObjects(bucketName, entry.Key)
				} else {
					showFileContent(bucketName, entry.Key, objectFlex)
				}
			}
		})

		// Update path display when selection changes
		objectTable.SetSelectionChangedFunc(func(row, column int) {
			if row > 0 && row-1 < len(objectEntries) { // Skip header row
				filename := objectEntries[row-1].Key
				path := fmt.Sprintf("s3://%s/%s", bucketName, filename)
				text.SetText(path)
			}
		})

		// Function to populate the table with current data
		populateObjectTable := func() {
			objectTable.Clear()
			objectEntries = nil // Reset entries

			go func() {
				objects, err := listS3Objects(context.TODO(), client, bucketName, prefix)
				if err != nil {
					log.Printf("failed to list objects: %v", err)
					return
				}

				app.QueueUpdateDraw(func() {
					// Add table headers
					objectTable.SetCell(0, 0, tview.NewTableCell("Name").SetTextColor(tcell.ColorYellow).SetSelectable(false))
					objectTable.SetCell(0, 1, tview.NewTableCell("Size").SetTextColor(tcell.ColorYellow).SetSelectable(false))
					objectTable.SetCell(0, 2, tview.NewTableCell("Modified").SetTextColor(tcell.ColorYellow).SetSelectable(false))

					row := 1

					// Add directories first
					for _, p := range objects.CommonPrefixes {
						entry := ObjectEntry{
							Key:         *p.Prefix,
							IsDirectory: true,
						}
						objectEntries = append(objectEntries, entry)
						objectTable.SetCell(row, 0, tview.NewTableCell(*p.Prefix).SetTextColor(tcell.ColorBlue))
						objectTable.SetCell(row, 1, tview.NewTableCell("DIR").SetTextColor(tcell.ColorBlue))
						objectTable.SetCell(row, 2, tview.NewTableCell("").SetTextColor(tcell.ColorBlue))
						row++
					}

					// Add files
					for _, o := range objects.Contents {
						if *o.Key != prefix {
							entry := ObjectEntry{
								Key:          *o.Key,
								IsDirectory:  false,
								Size:         *o.Size,
								LastModified: o.LastModified,
							}
							objectEntries = append(objectEntries, entry)

							sizeStr := formatFileSize(*o.Size)
							dateStr := formatDate(o.LastModified)

							objectTable.SetCell(row, 0, tview.NewTableCell(*o.Key))
							objectTable.SetCell(row, 1, tview.NewTableCell(sizeStr))
							objectTable.SetCell(row, 2, tview.NewTableCell(dateStr))
							row++
						}
					}

					// Select first data row if available
					if row > 1 {
						objectTable.Select(1, 0)
					}
				})
			}()
		}

		// Set this as the current refresh function for resize handling
		currentRefreshFunc = populateObjectTable

		// Set up input capture for the object table
		objectTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// Handle refresh to update formatting when terminal is resized
			if event.Key() == tcell.KeyCtrlL {
				populateObjectTable()
				return nil
			}
			// Handle existing navigation logic
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
				row, _ := objectTable.GetSelection()
				if row > 0 && row-1 < len(objectEntries) { // Skip header row
					entry := objectEntries[row-1]
					if entry.IsDirectory {
						listObjects(bucketName, entry.Key)
					} else {
						showFileContent(bucketName, entry.Key, objectFlex)
					}
				}
				return nil
			}
			return event
		})

		app.SetRoot(objectFlex, true)
		populateObjectTable()
	}

	bucketTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyRight {
			row, _ := bucketTable.GetSelection()
			if row > 0 && row-1 < len(bucketEntries) { // Skip header row
				bucketName := *bucketEntries[row-1].Name
				listObjects(bucketName, "")
			}
		}
		return event
	})

	// Function to print current URL on exit
	printCurrentURL := func() {
		if currentState.CurrentBucket != "" {
			url := fmt.Sprintf("s3://%s", currentState.CurrentBucket)
			if currentState.CurrentPrefix != "" {
				url += "/" + currentState.CurrentPrefix
			}
			fmt.Println(url)
		}
	}

	// Keybindings
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			printCurrentURL()
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			printCurrentURL()
			app.Stop()
			return nil
		}
		// Handle global refresh for window resize (Ctrl+L)
		if event.Key() == tcell.KeyCtrlL && currentRefreshFunc != nil {
			currentRefreshFunc()
			return nil
		}
		return event
	})

	// Handle navigation: URL argument takes precedence over saved state
	var targetBucket, targetPrefix string
	var shouldNavigate bool

	if s3URL != "" {
		// URL argument provided - use it
		targetBucket = currentState.CurrentBucket
		targetPrefix = currentState.CurrentPrefix
		shouldNavigate = true
	} else if savedState.CurrentBucket != "" {
		// No URL argument but saved state exists - use saved state
		targetBucket = savedState.CurrentBucket
		targetPrefix = savedState.CurrentPrefix
		shouldNavigate = true
	}

	if shouldNavigate {
		go func() {
			// Wait for buckets to be loaded first
			buckets, err := getBuckets(context.TODO(), client)
			if err != nil {
				return
			}

			// Check if the bucket exists
			bucketExists := false
			for _, bucket := range buckets {
				if *bucket.Name == targetBucket {
					bucketExists = true
					break
				}
			}

			if bucketExists {
				app.QueueUpdateDraw(func() {
					listObjects(targetBucket, targetPrefix)
				})
			}
		}()
	}

	// Run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print URL on normal exit
	printCurrentURL()
}
