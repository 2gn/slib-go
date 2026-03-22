// Package scraper provides functions to scrape book information from the SIT library website.
package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/2gn/slib-go/models"
	"github.com/PuerkitoBio/goquery"
)

var (
	spaceRegex = regexp.MustCompile(`\s+`)
	numRegex   = regexp.MustCompile(`^\d+\.\s*`)
)

// Scraper provides methods to fetch and parse book data.
type Scraper struct {
	BaseURL       string
	OnlineBaseURL string
	Campus        string // "Toyosu" or "Omiya"
}

// NewScraper creates and returns a new Scraper instance with default URLs.
func NewScraper() *Scraper {
	return &Scraper{
		BaseURL:       "https://slib.shibaura-it.ac.jp/sublib/ja/nalis_sl/display_panel",
		OnlineBaseURL: "https://slib.shibaura-it.ac.jp/sublib/ja/nalis_sl/display_slPanel",
	}
}

// GetDetailByID fetches detailed information about a book using its BIB ID.
func (s *Scraper) GetDetailByID(id string) (*models.BookDetail, error) {
	if id == "" {
		return nil, fmt.Errorf("empty book ID")
	}
	bookURL := fmt.Sprintf("https://library.shibaura-it.ac.jp/opc//recordID/catalog.bib/%s", id)
	return s.GetDetail(bookURL)
}

// GetDetail fetches detailed information about a book.
func (s *Scraper) GetDetail(bookURL string) (*models.BookDetail, error) {
	if bookURL == "" {
		return nil, fmt.Errorf("empty book URL")
	}

	resp, err := http.Get(bookURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	detail := &models.BookDetail{}
	detail.Title = s.normalizeSpace(doc.Find(".mainBox h3").First().Text())

	// Scrape Google Books URL and Image from bib page
	detail.GoogleBooksURL, _ = doc.Find("#xc-search-full-left > a:nth-child(2)").Attr("href")
	detail.ImageURL, _ = doc.Find("#xc-search-full-left img").Attr("src")

	doc.Find(".mainTable dt").Each(func(_ int, sel *goquery.Selection) {
		label := strings.TrimSpace(sel.Text())
		value := s.normalizeSpace(sel.Next().Text())

		switch label {
		case "フォーマット:":
			detail.Format = value
		case "責任表示:", "著者名:":
			if detail.Author == "" {
				detail.Author = value
			}
		case "言語:":
			detail.Language = value
		case "出版情報:":
			detail.Publication = value
		case "形態:":
			detail.PhysicalDesc = value
		case "ISBN:":
			detail.ISBN = value
		case "書誌ID:":
			detail.BibID = value
		}
	})

	// Fetch holdings via simplified AJAX
	if detail.BibID != "" {
		holdings, err := s.fetchHoldings(detail.BibID)
		if err == nil {
			detail.Holdings = holdings
		}
	}

	return detail, nil
}

func (s *Scraper) fetchHoldings(bibID string) ([]models.Holding, error) {
	ajaxURL := fmt.Sprintf(
		"https://library.shibaura-it.ac.jp/opc/xc_search/ajax/ncip_info_full?provider_id=1&bib_ids=%s",
		bibID,
	)

	resp, err := http.Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var jsonResp struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResp.Content))
	if err != nil {
		return nil, err
	}

	var holdings []models.Holding
	doc.Find("tbody tr").Each(func(_ int, sel *goquery.Selection) {
		holding := models.Holding{}

		// Status: .bkAva dd
		holding.Status = s.normalizeSpace(sel.Find(".bkAva dd").Text())

		// Location: .bkLoc dd
		holding.Location = s.normalizeSpace(sel.Find(".bkLoc dd").Text())

		// Call No: .bkCnu dd (use spDisInl for cleaner text)
		holding.CallNo = s.normalizeSpace(sel.Find(".bkCnu dd .spDisInl").Text())
		if holding.CallNo == "" {
			holding.CallNo = s.normalizeSpace(sel.Find(".bkCnu dd").Text())
		}

		if holding.Location != "" {
			holdings = append(holdings, holding)
		}
	})

	return holdings, nil
}

// ScrapeAll fetches both library and online publications concurrently.
func (s *Scraper) ScrapeAll(query string) ([]models.Book, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allBooks []models.Book
	var errs []string

	wg.Add(2)

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

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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

	doc.Find("li.d-flex.align-items-center.mt-2.border.rounded.p-2").Each(func(_ int, selection *goquery.Selection) {
		book := models.Book{IsOnline: isOnline}

		spanMe3 := selection.Find("span.me-3")

		link := spanMe3.Find("a.link-black")
		if link.Length() > 0 {
			book.URL, _ = link.Attr("href")
			book.Title = s.normalizeSpace(link.Text())
		}

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
