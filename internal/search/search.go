// Package search is vibe-doc's in-memory BM25 inverted index. Spec §8.
//
// Swap-path contract: a drop-in replacement (e.g. wrapping
// github.com/blevesearch/bleve/v2) must satisfy this minimal interface —
// nothing else in the codebase touches the internals:
//
//	type Searcher interface {
//	    AddDoc(d IndexedDoc)
//	    Search(q string) []Result
//	    SearchWithTimeout(q string, dur time.Duration) ([]Result, bool)
//	}
//
// To swap to Bleve, implement the methods and replace `New()` in
// server.go with the alternate constructor. Spec ref: §13.1.
package search

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Result is one search hit.
type Result struct {
	DocID   string  `json:"-"`
	URL     string  `json:"url"`
	File    string  `json:"file"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}

// HeadingHit is one heading in an indexed doc (Level + Text).
type HeadingHit struct {
	Level int
	Text  string
}

// IndexedDoc is the input passed to AddDoc when populating the index.
// Headings and Tags get extra weight per §8.
type IndexedDoc struct {
	ID       string
	URL      string
	File     string
	Title    string
	Body     string
	Headings []HeadingHit
	Tags     []string
}

// doc is the internal per-doc record.
type doc struct {
	id     string
	url    string
	file   string
	title  string
	body   string
	terms  map[string]float64 // weighted term frequency
	length int                // total weighted terms
}

// Index is a thread-safe inverted index.
type Index struct {
	mu     sync.RWMutex
	docs   map[string]*doc
	termDF map[string]int // document frequency per term
	avgLen float64
}

const (
	bm25K1     = 1.2
	bm25B      = 0.75
	maxResults = 50
)

// New returns an empty Index.
func New() *Index {
	return &Index{
		docs:   map[string]*doc{},
		termDF: map[string]int{},
	}
}

// AddDoc inserts (or replaces) a document. Weighting per spec §8:
// H1 ×3, H2 ×2, H3 ×1.5, tags ×2, body ×1.
func (i *Index) AddDoc(d IndexedDoc) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if old, ok := i.docs[d.ID]; ok {
		for t := range old.terms {
			i.termDF[t]--
			if i.termDF[t] <= 0 {
				delete(i.termDF, t)
			}
		}
	}
	rec := &doc{
		id:    d.ID,
		url:   d.URL,
		file:  d.File,
		title: d.Title,
		body:  d.Body,
		terms: map[string]float64{},
	}
	addStream := func(text string, weight float64) {
		for _, t := range tokenize(text) {
			rec.terms[t] += weight
			rec.length++
		}
	}
	// Title is treated as H1-equivalent.
	addStream(d.Title, 3.0)
	for _, h := range d.Headings {
		w := 1.0
		switch h.Level {
		case 1:
			w = 3.0
		case 2:
			w = 2.0
		case 3:
			w = 1.5
		}
		addStream(h.Text, w)
	}
	for _, tag := range d.Tags {
		addStream(tag, 2.0)
	}
	addStream(d.Body, 1.0)
	for t := range rec.terms {
		i.termDF[t]++
	}
	i.docs[d.ID] = rec
	i.recomputeAvgLen()
}

// Add is a back-compat helper: index id+title+body with no headings/tags.
func (i *Index) Add(id, title, body string) {
	i.AddDoc(IndexedDoc{ID: id, URL: "/" + id, File: id + ".md", Title: title, Body: body})
}

// AddWithURL indexes id+url+file+title+body with no headings/tags.
func (i *Index) AddWithURL(id, url, file, title, body string) {
	i.AddDoc(IndexedDoc{ID: id, URL: url, File: file, Title: title, Body: body})
}

// Remove drops a doc from the index.
func (i *Index) Remove(id string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	old, ok := i.docs[id]
	if !ok {
		return
	}
	for t := range old.terms {
		i.termDF[t]--
		if i.termDF[t] <= 0 {
			delete(i.termDF, t)
		}
	}
	delete(i.docs, id)
	i.recomputeAvgLen()
}

// Search returns up to maxResults hits ranked by BM25.
func (i *Index) Search(q string) []Result {
	i.mu.RLock()
	defer i.mu.RUnlock()
	terms, phrase := parseQuery(q)
	if len(terms) == 0 && phrase == "" {
		return nil
	}
	N := len(i.docs)
	if N == 0 {
		return nil
	}
	scores := map[string]float64{}
	for _, t := range terms {
		df := i.termDF[t]
		if df == 0 {
			continue
		}
		idf := math.Log(1 + (float64(N)-float64(df)+0.5)/(float64(df)+0.5))
		for _, d := range i.docs {
			tf := d.terms[t]
			if tf == 0 {
				continue
			}
			num := tf * (bm25K1 + 1)
			den := tf + bm25K1*(1-bm25B+bm25B*float64(d.length)/i.avgLen)
			scores[d.id] += idf * (num / den)
		}
	}
	if phrase != "" {
		for id, d := range i.docs {
			if strings.Contains(strings.ToLower(d.title+" "+d.body), phrase) {
				scores[id] += 5.0
			} else {
				delete(scores, id)
			}
		}
	}
	var ranked []Result
	for id, sc := range scores {
		d := i.docs[id]
		ranked = append(ranked, Result{
			DocID:   id,
			URL:     d.url,
			File:    d.file,
			Title:   d.title,
			Snippet: snippet(d.body, terms),
			Score:   sc,
		})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	if len(ranked) > maxResults {
		ranked = ranked[:maxResults]
	}
	return ranked
}

// SearchWithTimeout applies a hard cap so a runaway query can't block the
// handler. Returns (results, truncated). Truncated=true means the timeout
// fired before Search returned.
func (i *Index) SearchWithTimeout(q string, timeout time.Duration) ([]Result, bool) {
	type ret struct{ r []Result }
	out := make(chan ret, 1)
	go func() {
		out <- ret{r: i.Search(q)}
	}()
	select {
	case v := <-out:
		return v.r, false
	case <-time.After(timeout):
		return nil, true
	}
}

func (i *Index) recomputeAvgLen() {
	sum := 0
	for _, d := range i.docs {
		sum += d.length
	}
	if len(i.docs) > 0 {
		i.avgLen = float64(sum) / float64(len(i.docs))
	}
}

var (
	tokenRe  = regexp.MustCompile(`[A-Za-z][A-Za-z0-9\-]*`)
	phraseRe = regexp.MustCompile(`"([^"]+)"`)
)

func tokenize(s string) []string {
	matches := tokenRe.FindAllString(strings.ToLower(s), -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if _, stop := stopwords[m]; stop {
			continue
		}
		out = append(out, m)
	}
	return out
}

func parseQuery(q string) (terms []string, phrase string) {
	if m := phraseRe.FindStringSubmatch(q); len(m) == 2 {
		phrase = strings.ToLower(m[1])
		q = phraseRe.ReplaceAllString(q, " ")
	}
	terms = tokenize(q)
	return
}

func snippet(body string, terms []string) string {
	low := strings.ToLower(body)
	idx := -1
	for _, t := range terms {
		if j := strings.Index(low, t); j >= 0 {
			if idx < 0 || j < idx {
				idx = j
			}
		}
	}
	if idx < 0 {
		return truncate(body, 150)
	}
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := start + 150
	if end > len(body) {
		end = len(body)
	}
	snip := body[start:end]
	for _, t := range terms {
		snip = highlightCaseInsensitive(snip, t)
	}
	if start > 0 {
		snip = "…" + snip
	}
	if end < len(body) {
		snip = snip + "…"
	}
	return snip
}

func highlightCaseInsensitive(s, t string) string {
	low := strings.ToLower(s)
	out := strings.Builder{}
	last := 0
	tl := strings.ToLower(t)
	for {
		i := strings.Index(low[last:], tl)
		if i < 0 {
			out.WriteString(s[last:])
			break
		}
		i += last
		out.WriteString(s[last:i])
		out.WriteString("<mark>")
		out.WriteString(s[i : i+len(t)])
		out.WriteString("</mark>")
		last = i + len(t)
	}
	return out.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
