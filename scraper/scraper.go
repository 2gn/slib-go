package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/2gn/slib-go/models"
)

var (
	spaceRegex = regexp.MustCompile(`\s+`)
	numRegex   = regexp.MustCompile(`^\d+\.\s*`)
)

type Scraper struct {
	BaseURL       string
	OnlineBaseURL string
	Campus        string // "Toyosu" or "Omiya"
}

func NewScraper() *Scraper {
	return &Scraper{
		BaseURL:       "https://slib.shibaura-it.ac.jp/sublib/ja/nalis_sl/display_panel",
		OnlineBaseURL: "https://slib.shibaura-it.ac.jp/sublib/ja/nalis_sl/display_slPanel",
	}
}

// ScrapeAll fetches both library and online publications concurrently.
func (s *Scraper) ScrapeAll(query string) ([]models.Book, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allBooks []models.Book
	var errs []string

	wg.Add(2)

	// Fetch Library Collection
	go func() {
		defer wg.Done()
		books, err := s.Scrape(query)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Sprintf("library: %v", err))
			return
		}
		allBooks = append(allBooks, books...)
	}()

	// Fetch Online Publications
	go func() {
		defer wg.Done()
		books, err := s.ScrapeOnline(query)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			errs = append(errs, fmt.Sprintf("online: %v", err))
			return
		}
		allBooks = append(allBooks, books...)
	}()

	wg.Wait()

	if len(errs) > 0 {
		return allBooks, fmt.Errorf("errors occurred during concurrent scraping: %s", strings.Join(errs, "; "))
	}

	return allBooks, nil
}

// Scrape fetches and parses traditional library books (Panel 1)
func (s *Scraper) Scrape(query string) ([]models.Book, error) {
	url := fmt.Sprintf("%s?searchTarget=0&kw=%s&selectedLngOnly=0&selectSubject=1", s.BaseURL, query)
	if s.Campus != "" {
		url += fmt.Sprintf("&panel-1-facet-Library001=%s", s.Campus)
	}
	return s.fetchAndParse(url, false)
}

// ScrapeOnline fetches and parses online publications (Panel 2)
func (s *Scraper) ScrapeOnline(query string) ([]models.Book, error) {
	url := fmt.Sprintf("%s?searchTarget=0&kw=%s&selectedLngOnly=0&selectSubject=1&panelNo=2", s.OnlineBaseURL, query)
	// Online publications typically don't have a campus facet, but we keep the structure
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var jsonResp struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResp.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from JSON: %w", err)
	}

	return s.parse(doc, true), nil
}

func (s *Scraper) fetchAndParse(url string, isOnline bool) ([]models.Book, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return s.parse(doc, isOnline), nil
}

func (s *Scraper) parse(doc *goquery.Document, isOnline bool) []models.Book {
	var books []models.Book

	doc.Find("li.d-flex.align-items-center.mt-2.border.rounded.p-2").Each(func(i int, selection *goquery.Selection) {
		book := models.Book{IsOnline: isOnline}

		spanMe3 := selection.Find("span.me-3")
		
		// 1. URL and Title from Link
		link := spanMe3.Find("a.link-black")
		if link.Length() > 0 {
			book.URL, _ = link.Attr("href")
			book.Title = s.normalizeSpace(link.Text())
		}

		// 2. Author and fallback Title
		content := spanMe3.Clone()
		content.Find("a").Remove()
		
		fullText := s.normalizeSpace(content.Text())
		fullText = numRegex.ReplaceAllString(fullText, "")

		if book.Title != "" {
			book.Author = strings.TrimSpace(fullText)
		} else {
			book.Title = fullText
			book.Author = ""
		}

		// 3. Material
		book.Material = s.normalizeSpace(selection.Find("span.ms-auto").Text())

		books = append(books, book)
	})

	return books
}

func (s *Scraper) normalizeSpace(input string) string {
	input = strings.ReplaceAll(input, "\u00a0", " ")
	input = spaceRegex.ReplaceAllString(input, " ")
	return strings.TrimSpace(input)
}
