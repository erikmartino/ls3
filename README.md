# ls3 - S3 TUI Browser

A simple terminal-based user interface for browsing Amazon S3 buckets and objects.

## Features

- List and browse S3 buckets.
- Navigate through objects and folders within buckets.
- View text file content in full screen.

## Keybindings

| Key | Action |
|---|---|
| `Up/Down` | Navigate through lists |
| `Enter/Right` | Enter a bucket or folder |
| `Left` | Go back to the previous folder or bucket list |
| `q` | Quit the application |
| `Esc` | Go back from the file view |
| `Ctrl-C` | Quit the application |

## Build and Run

1.  Make sure you have Go installed and configured.
2.  Make sure you have your AWS credentials configured (e.g., via `~/.aws/credentials` or environment variables).
3.  Build the application:
    ```sh
    go build
    ```
4.  Run the application:
    ```sh
    ./ls3
    ```
