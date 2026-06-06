package mcp

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
)

// HandleSitemap serves a Google-News-compatible XML sitemap containing one
// <url> entry per public MCP-enabled function. Used by the marketing
// site (via XHR / direct link from the registry landing page) so that
// every per-function page is indexable.
//
// Route: GET /v1/mcp/sitemap.xml
//
// Pagination: ?page=N&size=1000 (max 1000 URLs per file, per the
// sitemap.org spec). Callers should generate a sitemap index at
// /v1/mcp/sitemap-index.xml and link to per-page files.
func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		http.Error(w, "MCP registry is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	page := atoiDefault(r.URL.Query().Get("page"), 0)
	size := atoiDefault(r.URL.Query().Get("size"), 1000)
	if size <= 0 || size > 50000 {
		size = 1000
	}
	offset := page * size

	rows, _, err := h.Store.ListEnabledMCPSettings(r.Context(), "", "", 0, size, offset)
	if err != nil {
		http.Error(w, "failed to list", http.StatusInternalServerError)
		return
	}

	// Resolve function rows so we have author/name + updated_at.
	type entry struct {
		loc        string
		lastmod    string
		changefreq string
		priority   string
	}
	out := make([]entry, 0, len(rows))
	base := h.publicURL()
	for _, row := range rows {
		fn, err := h.Store.GetFunctionByID(r.Context(), row.FunctionID)
		if err != nil || fn == nil {
			continue
		}
		loc := fmt.Sprintf("%s/@%s/v1/fx/%s", base, fn.Author, fn.Name)
		lastmod := row.UpdatedAt.UTC().Format("2006-01-02")
		out = append(out, entry{
			loc:        loc,
			lastmod:    lastmod,
			changefreq: "daily",
			priority:   "0.8",
		})
	}

	// Render XML.
	type xmlURL struct {
		XMLName    xml.Name `xml:"url"`
		Loc        string   `xml:"loc"`
		LastMod    string   `xml:"lastmod,omitempty"`
		ChangeFreq string   `xml:"changefreq,omitempty"`
		Priority   string   `xml:"priority,omitempty"`
	}
	type urlset struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		URLs    []xmlURL `xml:"url"`
	}
	urls := urlset{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]xmlURL, 0, len(out)),
	}
	for _, e := range out {
		urls.URLs = append(urls.URLs, xmlURL{
			Loc:        e.loc,
			LastMod:    e.lastmod,
			ChangeFreq: e.changefreq,
			Priority:   e.priority,
		})
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ToolsIndexCacheMaxAge))
	setCORSHeaders(w, r)
	_, _ = w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(urls)
}

// HandleSitemapIndex returns a sitemap index pointing to per-page sitemaps.
// The total page count is determined by counting enabled MCP functions and
// dividing by the page size. Cached for the same duration as the index.
func (h *Handler) HandleSitemapIndex(w http.ResponseWriter, r *http.Request) {
	if h.Disabled {
		http.Error(w, "MCP registry is temporarily unavailable", http.StatusServiceUnavailable)
		return
	}
	size := atoiDefault(r.URL.Query().Get("size"), 1000)
	if size <= 0 || size > 50000 {
		size = 1000
	}
	_, total, err := h.Store.ListEnabledMCPSettings(r.Context(), "", "", 0, 1, 0)
	_ = total
	if err != nil {
		http.Error(w, "failed to count", http.StatusInternalServerError)
		return
	}
	// Cheap full count via a separate call (limit = 1, offset = 0 returns total).
	rows, total2, _ := h.Store.ListEnabledMCPSettings(r.Context(), "", "", 0, 1, 0)
	_ = rows
	total = total2

	pages := (total + size - 1) / size
	if pages < 1 {
		pages = 1
	}
	base := h.publicURL()

	type sitemap struct {
		XMLName xml.Name `xml:"sitemap"`
		Loc     string   `xml:"loc"`
		LastMod string   `xml:"lastmod,omitempty"`
	}
	type indexSet struct {
		XMLName   xml.Name  `xml:"sitemapindex"`
		XMLNS     string    `xml:"xmlns,attr"`
		Sitemaps  []sitemap `xml:"sitemap"`
	}
	out := indexSet{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	today := time.Now().UTC().Format("2006-01-02")
	for i := 0; i < pages; i++ {
		out.Sitemaps = append(out.Sitemaps, sitemap{
			Loc:     fmt.Sprintf("%s/v1/mcp/sitemap.xml?page=%d&size=%d", base, i, size),
			LastMod: today,
		})
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ToolsIndexCacheMaxAge))
	setCORSHeaders(w, r)
	_, _ = w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	_ = enc.Encode(out)
}

// atoiDefault parses s as int, returning def if s is empty/invalid.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var n int
	for _, r := range strings.TrimSpace(s) {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// _ keeps registry referenced in case future sitemap enrichment needs it
// (e.g. per-function changefreq or priority).
var _ = registry.MCPSettings{}
var _ = mux.Vars
