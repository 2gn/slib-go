package scraper

import (
	"testing"
)

func TestScrape(t *testing.T) {
	s := NewScraper()
	books, err := s.Scrape("test")
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(books) == 0 {
		t.Error("Expected at least one book, got 0")
	}

	for _, book := range books {
		if book.Title == "" {
			t.Error("Found book with empty title")
		}
		if book.URL == "" {
			t.Error("Found book with empty URL")
		}
	}
}
