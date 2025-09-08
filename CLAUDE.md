# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ls3 is a terminal-based user interface (TUI) for browsing Amazon S3 buckets and objects. It's built in Go using the tview library for the UI and AWS SDK v2 for S3 operations.

## Commands

### Building and Running
```bash
# Build the application
go build

# Run the application
./ls3

# Run with specific S3 URL
./ls3 s3://bucket-name/prefix/
```

### Development
```bash
# Run tests
go test ./...

# Run specific test
go test -run TestFunctionName

# Vet code for issues  
go vet ./...

# Clean and verify modules
go mod tidy
go mod verify
```

### Release (using goreleaser)
```bash
# Test release build locally
goreleaser build --snapshot --clean

# Create release (requires proper git tags and GitHub setup)
goreleaser release --clean
```

## Architecture

### Core Components

**main.go**: Contains the main application logic including:
- TUI setup using tview (list, table, text views)
- Application state management (current bucket/prefix with persistence)
- Keyboard navigation and event handling
- S3 URL parsing and argument handling
- File content viewing with gzip decompression support

**s3_client.go**: AWS S3 client abstraction with interface for testing:
- `S3Client` interface defining ListBuckets, ListObjectsV2, GetObject operations
- Wrapper functions: `getBuckets()`, `listS3Objects()`, `getObjectContent()`

**s3_client_test.go**: Unit tests with mock S3 client implementation

### Key Data Structures

- `AppState`: Tracks current bucket and prefix for session persistence
- `ObjectEntry`: Represents S3 objects/directories with metadata for display
- Mock client pattern for testing S3 operations

### State Management

The application persists the current navigation state (bucket/prefix) to `~/.ls3_state.json` and can restore the last browsed location on startup. Command-line S3 URLs take precedence over saved state.

### UI Flow

1. Bucket list view (entry point)
2. Object/folder navigation within buckets (table view)
3. File content display (text view) with automatic gzip decompression
4. Breadcrumb navigation and keyboard shortcuts

## Dependencies

- AWS SDK Go v2 for S3 operations
- tview for terminal UI components  
- tcell for terminal control
- Standard library for file operations, JSON state, gzip handling