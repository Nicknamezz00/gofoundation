// Package pagination provides shared cursor and offset pagination helpers.
package pagination

// Page holds offset-based pagination parameters.
type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PagedResult wraps a slice of items with total count metadata.
type PagedResult[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// DefaultPage returns a Page with sensible defaults.
func DefaultPage() Page {
	return Page{Limit: 20, Offset: 0}
}

// Clamp ensures limit stays within [1, maxLimit].
func (p Page) Clamp(maxLimit int) Page {
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > maxLimit {
		p.Limit = maxLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}
