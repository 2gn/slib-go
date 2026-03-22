package scraper

import (
	"testing"
)

func TestScrape(t *testing.T) {
	s := NewScraper()
	books, err := s.Scrape("Python")
	if err != nil {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(books) == 0 {
		t.Skip("No books found for 'Golang', skipping further tests")
	}

	for _, book := range books {
		if book.Title == "" {
			t.Error("Found book with empty title")
		}
		if book.URL == "" {
			t.Error("Found book with empty URL")
		}
	}

	// Test GetDetail for the first book
	if len(books) > 0 {
		detail, err := s.GetDetail(books[0].URL)
		if err != nil {
			t.Errorf("GetDetail failed for %s: %v", books[0].URL, err)
		} else {
			if detail.Title == "" {
				t.Error("Detail Title is empty")
			}
			// Note: GoogleBooksURL might be empty for some books
			t.Logf("Found Google Books URL: %s", detail.GoogleBooksURL)
		}
	}
}
