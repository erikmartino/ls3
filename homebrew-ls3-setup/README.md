# Homebrew Tap for ls3

This is a custom Homebrew tap for the ls3 S3 TUI browser.

## Installation

```bash
brew tap erikmartino/ls3
brew install --cask ls3
```

## About ls3

ls3 is a terminal-based user interface for browsing Amazon S3 buckets and objects built in Go. It provides an intuitive way to navigate S3 buckets and view file contents directly from your terminal.

### Features

- List and browse S3 buckets
- Navigate through objects and folders within buckets
- View text file content in full screen
- Automatic gzip decompression for compressed files
- Session persistence (remembers your last browsed location)

### Requirements

- AWS credentials configured (via `~/.aws/credentials` or environment variables)
- macOS (via Homebrew)

## Manual Installation

If you prefer to install manually, you can download the latest release from the [releases page](https://github.com/erikmartino/ls3/releases).

## Issues and Contributions

Please report issues or contribute to the main repository: https://github.com/erikmartino/ls3

## License

MIT License - see the [LICENSE](https://github.com/erikmartino/ls3/blob/main/LICENSE) file for details.