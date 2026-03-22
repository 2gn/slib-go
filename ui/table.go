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

// RenderDetailTable returns a rendered string of a transposed table containing book details and holdings.
func RenderDetailTable(d *models.BookDetail) string {
	// Combined Transposed Table
	numCols := 1 + len(d.Holdings)
	if numCols < 2 {
		numCols = 2 // At least Labels + Data
	}

	// Field labels
	fields := []string{"Title", "Author", "Publication", "ISBN", "Format"}
	if d.GoogleBooksURL != "" {
		fields = append(fields, "Google Books")
	}

	rows := [][]string{}
	// General Detail rows
	for _, f := range fields {
		row := make([]string, numCols)
		row[0] = f
		switch f {
		case "Title":
			row[1] = d.Title
		case "Author":
			row[1] = d.Author
		case "Publication":
			row[1] = d.Publication
		case "ISBN":
			row[1] = d.ISBN
		case "Format":
			row[1] = d.Format
		case "Google Books":
			row[1] = d.GoogleBooksURL
		}
		rows = append(rows, row)
	}

	// Transposed Holdings rows
	if len(d.Holdings) > 0 {
		holdingFields := []string{"Location", "Call No", "Status"}
		for _, f := range holdingFields {
			row := make([]string, numCols)
			row[0] = f
			for i, h := range d.Holdings {
				switch f {
				case "Location":
					row[i+1] = h.Location
				case "Call No":
					row[i+1] = h.CallNo
				case "Status":
					row[i+1] = ColorizeStatus(h.Status)
				}
			}
			rows = append(rows, row)
		}
	}

	t := lTable.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Rows(rows...)

	return t.Render()
}
