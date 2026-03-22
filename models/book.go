package models

type Book struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	URL      string `json:"url"`
	Material string `json:"material"`
	IsOnline bool   `json:"is_online"`
}

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

type Holding struct {
	Location string `json:"location"`
	Status   string `json:"status"`
	CallNo   string `json:"call_no"`
}
