// Package models defines the data structures used by the SIT search scraper.
package models

// Book represents a book item in the search results.
type Book struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	URL      string `json:"url"`
	Material string `json:"material"`
	IsOnline bool   `json:"is_online"`
}

// BookDetail represents the detailed information about a specific book.
type BookDetail struct {
	Title          string    `json:"title"`
	Format         string    `json:"format"`
	Author         string    `json:"author"`
	Language       string    `json:"language"`
	Publication    string    `json:"publication"`
	PhysicalDesc   string    `json:"physical_desc"`
	ISBN           string    `json:"isbn"`
	BibID          string    `json:"bib_id"`
	GoogleBooksURL string    `json:"google_books_url"`
	ImageURL       string    `json:"image_url"`
	Holdings       []Holding `json:"holdings"`
}

// Holding represents the physical location and status of a book in the library.
type Holding struct {
	Location string `json:"location"`
	Status   string `json:"status"`
	CallNo   string `json:"call_no"`
}
