// Package sitemap generates a /sitemap.xml from the walked mount tree.
package sitemap

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/genno-whittlery/vibe-doc/internal/mount"
	"github.com/genno-whittlery/vibe-doc/internal/walk"
)

type urlEntry struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

type urlset struct {
	XMLName xml.Name   `xml:"urlset"`
	XMLNS   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

// Generate writes a sitemap.xml document to w. baseURL is the scheme+host
// (e.g. "http://127.0.0.1:4000") without trailing slash. set provides the
// mount tree to walk.
func Generate(w io.Writer, baseURL string, set *mount.Set) error {
	us := urlset{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	baseURL = strings.TrimRight(baseURL, "/")
	for _, m := range set.Mounts() {
		files, err := walk.MD(m.Root)
		if err != nil {
			continue
		}
		for _, f := range files {
			stem := trimMd(f)
			// README → folder URL
			if strings.HasSuffix(stem, "/README") {
				stem = strings.TrimSuffix(stem, "README")
			} else if stem == "README" {
				stem = ""
			}
			var loc string
			if m.URL == "/" {
				loc = baseURL + "/" + stem
			} else {
				loc = baseURL + m.URL + "/" + stem
			}
			us.URLs = append(us.URLs, urlEntry{Loc: loc})
		}
	}
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(us)
}

func trimMd(s string) string {
	if strings.HasSuffix(s, ".md") {
		return s[:len(s)-3]
	}
	return s
}
