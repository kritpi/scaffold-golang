package domain

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Page is an offset page request. Number is 1-based.
type Page struct {
	Number int
	Size   int
}

// Normalize sets Number to at least 1, defaults Size to DefaultPageSize and caps it at MaxPageSize.
func (p Page) Normalize() Page {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = DefaultPageSize
	}
	if p.Size > MaxPageSize {
		p.Size = MaxPageSize
	}
	return p
}

// Offset is the number of items before page p.
func (p Page) Offset() int { return (p.Number - 1) * p.Size }

// PageResult is one page of items and the total count.
type PageResult[T any] struct {
	Items []T
	Page  Page
	Total int
}

// LastPage is the number of the last page, at least 1.
func (r PageResult[T]) LastPage() int {
	if r.Total == 0 || r.Page.Size == 0 {
		return 1
	}
	return (r.Total + r.Page.Size - 1) / r.Page.Size
}
