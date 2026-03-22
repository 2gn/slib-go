package models

type Book struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	URL      string `json:"url"`
	Material string `json:"material"`
	IsOnline bool   `json:"is_online"`
}
