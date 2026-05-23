// Package search builds an inverted index over markdown bodies and serves
// BM25 ranked queries. Spec ref: §8. This file is a STUB — Tasks 12-13
// replace it with the full implementation.
package search

// Heading is one heading row in an indexed doc (used for boost weighting).
type Heading struct {
	Text  string
	Level int
}

// IndexedDoc is the input passed to (*Index).AddDoc when populating the
// index. Tasks 12-13 use Headings + Tags for BM25 section boost.
type IndexedDoc struct {
	ID       string
	URL      string
	Title    string
	Body     string
	Headings []Heading
	Tags     []string
}

// Result is one search hit returned by (*Index).Search.
type Result struct {
	URL   string
	Title string
	Score float64
}

// Index is the in-memory search index. STUB: holds nothing. Task 12 adds
// the inverted-index + BM25 implementation.
type Index struct{}

// New returns an empty Index. STUB.
func New() *Index { return &Index{} }

// AddDoc indexes one doc. STUB: no-op.
func (i *Index) AddDoc(_ IndexedDoc) {}

// Search returns ranked hits for the query. STUB: returns nil.
func (i *Index) Search(_ string) []Result { return nil }
