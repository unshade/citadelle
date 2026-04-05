// Package pagination provides shared types for passing and returning pagination
// state across the repository, service, and handler layers.
package pagination

const (
	DefaultPage    uint64 = 1
	DefaultPerPage uint64 = 50
	MaxPerPage     uint64 = 100
)

// Params carries the caller's requested page and page size.
// Always call Normalize before using Offset/Limit in a DB query.
type Params struct {
	Page    uint64
	PerPage uint64
}

// Normalize returns a copy with defaults applied and PerPage capped at MaxPerPage.
func (p Params) Normalize() Params {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
	return p
}

// Offset returns the DB row offset for the current page (0-based).
func (p Params) Offset() int {
	n := p.Normalize()
	return int((n.Page - 1) * n.PerPage)
}

// Limit returns the DB row limit (= PerPage after normalization).
func (p Params) Limit() int {
	return int(p.Normalize().PerPage)
}

// Result is returned alongside paginated data.
// It echoes the effective Params and adds the Total row count so callers
// can compute the total number of pages without an extra round-trip.
type Result struct {
	Page    uint64 `json:"page"`
	PerPage uint64 `json:"perPage"`
	Total   uint64 `json:"total"`
}
