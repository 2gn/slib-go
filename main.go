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
)

func main() {
	mode := flag.String("mode", "all", "Search mode: all, library, online")
	toyosu := flag.Bool("toyosu", false, "Filter by Toyosu campus")
	omiya := flag.Bool("omiya", false, "Filter by Omiya campus")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "SIT Search Scraper\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go run main.go [flags] [query]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  go run main.go test\n")
		fmt.Fprintf(os.Stderr, "  go run main.go -omiya physics\n")
		fmt.Fprintf(os.Stderr, "  go run main.go -mode library -toyosu \"computer science\"\n")
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *toyosu && *omiya {
		log.Fatal("Error: Cannot specify both -toyosu and -omiya. Please choose one campus or none for all campuses.")
	}

	query := "test"
	if flag.NArg() > 0 {
		query = strings.Join(flag.Args(), " ")
	}

	s := scraper.NewScraper()
	
	if *toyosu {
		s.Campus = "Toyosu"
	} else if *omiya {
		s.Campus = "Omiya"
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
		log.Fatalf("Invalid mode: %s. Use 'all', 'library', or 'online'.", *mode)
	}

	if err != nil {
		if *jsonOutput {
			log.Fatalf("{\"error\": \"%v\"}", err)
		} else {
			log.Printf("Error during scraping: %v", err)
		}
	}

	if *jsonOutput {
		output, err := json.MarshalIndent(books, "", "  ")
		if err != nil {
			log.Fatalf("{\"error\": \"failed to marshal JSON: %v\"}", err)
		}
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
