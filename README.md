# slib-go

A powerful, Go-based crawler and terminal interface for the Shibaura Institute of Technology (SIT) library (https://slib.shibaura-it.ac.jp).

`slib-go` is designed for both **AI agents** and **human users**, providing both machine-readable JSON output and a rich, interactive Terminal User Interface (TUI).

## Key Features

- **Concurrent Scraping**: Search both physical library books and online publications (e-books, journals) simultaneously.
- **Campus Filtering**: Filter search results by campus (Toyosu, Omiya).
- **Detailed Book Info**: Fetch full bibliographic details including ISBN, publication info, and real-time holdings status.
- **AI-Friendly**: Output results in clean, structured JSON format for easy integration with AI agents and automation scripts.
- **Interactive TUI**: A modern, interactive terminal interface for searching, browsing, and viewing book details with ASCII art cover previews (where available).

## Quick Start

### For Humans: Interactive TUI

Launch the interactive interface to search and browse:

```bash
go run main.go tui
```

### For AI Agents & Automation: JSON Output

Get structured data for a specific book:

```bash
go run main.go -json -id BB16343390
```

## Documentation

Detailed usage instructions and examples can be found in [docs/usage.md](docs/usage.md).

## Compatibility

- Tested on `amd64 linux`.
- Requires Go 1.21+

## License

(MIT License - or check the repository for details)
