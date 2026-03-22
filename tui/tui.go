package tui

import (
	"fmt"

	"github.com/2gn/slib-go/models"
	"github.com/2gn/slib-go/scraper"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

func NewModel(s *scraper.Scraper) model {
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

	if m.searching {
		s += "Searching...\n"
	} else if m.err != nil {
		s += fmt.Sprintf("Error: %v\n", m.err)
	} else {
		s += baseStyle.Render(m.table.View()) + "\n"
	}

	s += "\n(enter: search/details, /: focus input, q: quit)\n"

	if m.loadingDetail {
		s += "\nLoading book details...\n"
	} else if m.detail != nil {
		d := m.detail
		detailText := fmt.Sprintf("\n--- Details ---\nTitle: %s\nAuthor: %s\nPublication: %s\nISBN: %s\nFormat: %s\n",
			d.Title, d.Author, d.Publication, d.ISBN, d.Format)

		if d.GoogleBooksURL != "" {
			detailText += fmt.Sprintf("Google Books: %s\n", d.GoogleBooksURL)
		}

		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(detailText)

		if len(d.Holdings) > 0 {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\n--- Holdings ---\n")
			for _, h := range d.Holdings {
				s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
					fmt.Sprintf("%s | %s | %s\n", h.Location, h.CallNo, h.Status))
			}
		}
	} else if len(m.books) > 0 && !m.searching {
		curr := m.table.Cursor()
		if curr >= 0 && curr < len(m.books) {
			b := m.books[curr]
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
				fmt.Sprintf("\nSelected: %s\nAuthor: %s\nPress Enter for more details\n", b.Title, b.Author))
		}
	}

	return s
}

func StartTUI(s *scraper.Scraper) error {
	p := tea.NewProgram(NewModel(s))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
