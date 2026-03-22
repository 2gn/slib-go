// Package tui provides a terminal user interface for the SIT search scraper.
package tui

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"net/http"

	"github.com/2gn/slib-go/models"
	"github.com/2gn/slib-go/scraper"
	"github.com/2gn/slib-go/ui"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qeesung/image2ascii/convert"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	textInput     textinput.Model
	table         table.Model
	scraper       *scraper.Scraper
	books         []models.Book
	detail        *models.BookDetail
	err           error
	searching     bool
	loadingDetail bool
	image         string
}

// NewModel creates and returns a new TUI model.
func NewModel(s *scraper.Scraper) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Search query..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	columns := []table.Column{
		{Title: "Type", Width: 10},
		{Title: "Title", Width: 40},
		{Title: "Author", Width: 30},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	sTable := table.DefaultStyles()
	sTable.Header = sTable.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	sTable.Selected = sTable.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(sTable)

	return model{
		textInput: ti,
		table:     t,
		scraper:   s,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

type searchMsg struct {
	books []models.Book
	err   error
}

type detailMsg struct {
	detail *models.BookDetail
	err    error
}

type imageMsg struct {
	image string
}

func (m model) updateSearch(query string) tea.Cmd {
	return func() tea.Msg {
		books, err := m.scraper.ScrapeAll(query)
		return searchMsg{books: books, err: err}
	}
}

func (m model) fetchDetail(url string) tea.Cmd {
	return func() tea.Msg {
		detail, err := m.scraper.GetDetail(url)
		return detailMsg{detail: detail, err: err}
	}
}

func (m model) fetchImage(url string, height int) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(url)
		if err != nil {
			return imageMsg{image: ""}
		}
		defer func() { _ = resp.Body.Close() }()

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return imageMsg{image: ""}
		}

		converter := convert.NewImageConverter()
		options := convert.DefaultOptions
		options.Ratio = 0.2
		options.FixedHeight = height
		options.FixedWidth = height * 2 // Roughly maintain aspect ratio in terminal (2x taller than wide characters)
		ascii := converter.Image2ASCIIString(img, &options)
		return imageMsg{image: ascii}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if !m.textInput.Focused() || msg.String() != "q" {
				return m, tea.Quit
			}
		case "/":
			if !m.textInput.Focused() {
				m.textInput.Focus()
				m.table.Blur()
				return m, nil
			}
		case "delete":
			m.textInput.SetValue("")
			m.textInput.Focus()
			m.table.Blur()
			return m, nil
		case "enter":
			if m.textInput.Focused() {
				m.searching = true
				m.books = nil
				m.detail = nil
				m.table.SetRows(nil)
				query := m.textInput.Value()
				m.textInput.Blur()
				m.table.Focus()
				return m, m.updateSearch(query)
			}
			if m.table.Focused() && len(m.books) > 0 {
				curr := m.table.Cursor()
				if curr >= 0 && curr < len(m.books) {
					book := m.books[curr]
					if book.URL != "" {
						m.loadingDetail = true
						m.detail = nil
						m.image = ""
						return m, m.fetchDetail(book.URL)
					}
				}
			}
		}

	case searchMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.books = msg.books
		var rows []table.Row
		for _, b := range m.books {
			loc := "Library"
			if b.IsOnline {
				loc = "Online"
			}
			rows = append(rows, table.Row{loc, b.Title, b.Author})
		}
		m.table.SetRows(rows)
		m.err = nil
		return m, nil

	case detailMsg:
		m.loadingDetail = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.detail = msg.detail
		if m.detail.ImageURL != "" {
			// Calculate detail height: 6 (base) + 1 (GoogleBooks)
			height := 6
			if m.detail.GoogleBooksURL != "" {
				height++
			}
			return m, m.fetchImage(m.detail.ImageURL, height)
		}
		return m, nil

	case imageMsg:
		m.image = msg.image
		return m, nil
	}

	if m.textInput.Focused() {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	var tCmd tea.Cmd
	m.table, tCmd = m.table.Update(msg)
	return m, tCmd
}

func (m model) View() string {
	s := "SIT Search TUI\n\n"
	s += m.textInput.View() + "\n\n"

	switch {
	case m.searching:
		s += "Searching...\n"
	case m.err != nil:
		s += fmt.Sprintf("Error: %v\n", m.err)
	default:
		s += baseStyle.Render(m.table.View()) + "\n"
	}

	s += "\n(enter: search/details, /: focus input, q: quit)\n"

	switch {
	case m.loadingDetail:
		s += "\nLoading book details...\n"
	case m.detail != nil:
		tableRendered := ui.RenderDetailTable(m.detail)

		if m.image != "" {
			s += "\n" + lipgloss.JoinHorizontal(lipgloss.Top, m.image, "  ", tableRendered) + "\n"
		} else {
			s += "\n" + tableRendered + "\n"
		}
	case len(m.books) > 0 && !m.searching:
		curr := m.table.Cursor()
		if curr >= 0 && curr < len(m.books) {
			b := m.books[curr]
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
				fmt.Sprintf("\nSelected: %s\nAuthor: %s\nPress Enter for more details\n", b.Title, b.Author))
		}
	}

	return s
}

// StartTUI starts the terminal user interface.
func StartTUI(s *scraper.Scraper) error {
	p := tea.NewProgram(NewModel(s))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
