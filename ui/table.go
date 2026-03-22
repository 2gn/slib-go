// Package ui provides shared UI components for CLI and TUI.
package ui

import (
	"strings"

	"github.com/2gn/slib-go/models"
	"github.com/charmbracelet/lipgloss"
	lTable "github.com/charmbracelet/lipgloss/table"
)

// ColorizeStatus returns a colorized string with an emoji for the given book status.
func ColorizeStatus(status string) string {
	switch {
	case strings.Contains(status, "貸出中"): // On Loan
		return lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Render("🕒 " + status)
	case strings.Contains(status, "利用可能"), strings.Contains(status, "在架"): // Available / On Shelf
		return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("✅ " + status)
	case strings.Contains(status, "予約中"): // Reserved
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("🔖 " + status)
	default:
		return status
	}
}

// RenderDetailTable returns a rendered string of two separate tables: book details and holdings.
// It tries to display them in a row if maxWidth allows, otherwise vertically.
func RenderDetailTable(d *models.BookDetail, maxWidth int) string {
	// Details Table
	detailRows := [][]string{
		{"Title", d.Title},
		{"Author", d.Author},
		{"Publication", d.Publication},
		{"ISBN", d.ISBN},
		{"Format", d.Format},
	}
	if d.GoogleBooksURL != "" {
		detailRows = append(detailRows, []string{"Google Books", d.GoogleBooksURL})
	}

	dt := lTable.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Rows(detailRows...)

	detailsRendered := dt.Render()

	// Holdings Table
	if len(d.Holdings) > 0 {
		// Horizontal header: Campus name (Location)
		headers := make([]string, len(d.Holdings)+1)
		headers[0] = ""
		for i, h := range d.Holdings {
			headers[i+1] = h.Location
		}

		// Row 1: Call No
		callNoRow := make([]string, len(d.Holdings)+1)
		callNoRow[0] = "Call No"
		for i, h := range d.Holdings {
			callNoRow[i+1] = h.CallNo
		}

		// Row 2: Status
		statusRow := make([]string, len(d.Holdings)+1)
		statusRow[0] = "Availability"
		for i, h := range d.Holdings {
			statusRow[i+1] = ColorizeStatus(h.Status)
		}

		ht := lTable.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
			Headers(headers...).
			Rows(callNoRow, statusRow)

		holdingsRendered := ht.Render()

		// Check if they fit in a row
		detailsWidth := lipgloss.Width(detailsRendered)
		holdingsWidth := lipgloss.Width(holdingsRendered)

		// If maxWidth is 0 (CLI mode with no terminal info or not specified), 

		// we can default to vertical or some large value.
		// For now, let's assume if maxWidth > 0, we check.
		if maxWidth > 0 && detailsWidth+holdingsWidth+4 <= maxWidth {
			return lipgloss.JoinHorizontal(lipgloss.Top, detailsRendered, "    ", holdingsRendered)
		}

		return lipgloss.JoinVertical(lipgloss.Left, detailsRendered, "\n"+holdingsRendered)
	}

	return detailsRendered
}
