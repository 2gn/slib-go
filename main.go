// Package main provides the entry point for the SIT search scraper CLI and TUI.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/2gn/slib-go/models"
	"github.com/2gn/slib-go/scraper"
	"github.com/2gn/slib-go/tui"
	"github.com/charmbracelet/lipgloss"
	lTable "github.com/charmbracelet/lipgloss/table"
)

func main() {
	mode := flag.String("mode", "all", "Search mode: all, library, online")
	toyosu := flag.Bool("toyosu", false, "Filter by Toyosu campus")
	omiya := flag.Bool("omiya", false, "Filter by Omiya campus")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	bookID := flag.String("id", "", "Fetch details for a specific book by ID (e.g., BB16343390)")
	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "SIT Search Scraper\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go run main.go [flags] [query]\n")
		fmt.Fprintf(os.Stderr, "  go run main.go -id [ID]\n")
		fmt.Fprintf(os.Stderr, "  go run main.go tui\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  go run main.go -id BB16343390\n")
		fmt.Fprintf(os.Stderr, "  go run main.go -mode library physics\n")
		fmt.Fprintf(os.Stderr, "  go run main.go tui\n")
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	s := scraper.NewScraper()

	// Direct ID lookup
	if *bookID != "" {
		detail, err := s.GetDetailByID(*bookID)
		if err != nil {
			if *jsonOutput {
				log.Fatalf("{\"error\": \"%v\"}", err)
			}
			log.Fatalf("Error fetching book details: %v", err)
		}

		if *jsonOutput {
			output, _ := json.MarshalIndent(detail, "", "  ")
			fmt.Println(string(output))
			return
		}

		rows := [][]string{
			{"Title", detail.Title},
			{"Author", detail.Author},
			{"Publication", detail.Publication},
			{"Format", detail.Format},
			{"ISBN", detail.ISBN},
			{"Bib ID", detail.BibID},
		}
		if detail.GoogleBooksURL != "" {
			rows = append(rows, []string{"Google Books", detail.GoogleBooksURL})
		}

		dt := lTable.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			Rows(rows...)

		fmt.Println(dt.Render())

		if len(detail.Holdings) > 0 {
			holdingRows := [][]string{}
			for _, h := range detail.Holdings {
				holdingRows = append(holdingRows, []string{
					h.Location,
					h.CallNo,
					colorizeStatus(h.Status),
				})
			}

			ht := lTable.New().
				Border(lipgloss.NormalBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
				Headers("Location", "Call No", "Status").
				Rows(holdingRows...)

			fmt.Println(ht.Render())
		}
		return
	}

	// TUI Mode
	isTui := false
	if flag.NArg() > 0 && flag.Arg(0) == "tui" {
		isTui = true
	} else if flag.NArg() == 0 && !isAnyFlagPresent() {
		isTui = true
	}

	if isTui {
		if err := tui.StartTUI(s); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
		return
	}

	// Search Mode
	if *toyosu {
		s.Campus = "Toyosu"
	} else if *omiya {
		s.Campus = "Omiya"
	}

	if *toyosu && *omiya {
		log.Fatal("Error: Cannot specify both -toyosu and -omiya.")
	}

	query := "test"
	if flag.NArg() > 0 {
		query = strings.Join(flag.Args(), " ")
	}

	var books []models.Book
	var err error

	if !*jsonOutput {
		campusStr := "All"
		if s.Campus != "" {
			campusStr = s.Campus
		}
		fmt.Printf("Searching for: %s (Mode: %s, Campus: %s)...\n", query, *mode, campusStr)
	}

	switch strings.ToLower(*mode) {
	case "library":
		books, err = s.Scrape(query)
	case "online":
		books, err = s.ScrapeOnline(query)
	case "all":
		books, err = s.ScrapeAll(query)
	default:
		log.Fatalf("Invalid mode: %s.", *mode)
	}

	if err != nil {
		if *jsonOutput {
			log.Fatalf("{\"error\": \"%v\"}", err)
		}
		log.Printf("Error during scraping: %v", err)
	}

	if *jsonOutput {
		output, _ := json.MarshalIndent(books, "", "  ")
		fmt.Println(string(output))
		return
	}

	fmt.Printf("Found %d items:\n", len(books))
	for i, book := range books {
		location := "Library"
		if book.IsOnline {
			location = "Online"
		}
		fmt.Printf("%d. [%s] %s\n   Author: %s\n   Material: %s\n   URL: %s\n\n",
			i+1, location, book.Title, book.Author, book.Material, book.URL)
	}
}

func isAnyFlagPresent() bool {
	var present bool
	flag.Visit(func(_ *flag.Flag) {
		present = true
	})
	return present
}

func colorizeStatus(status string) string {
	switch {
	case strings.Contains(status, "貸出中"): // On Loan
		return lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Render("🕒 " + status)
	case strings.Contains(status, "利用可能"), strings.Contains(status, "在架"): // Available / On Shelf
		return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("✅ " + status)
	case strings.Contains(status, "予約中"): // Reserved
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("🔖 " + status)
	default:
		return status
	}
}
