package note

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

//go:embed servedata/templates/*.html
var serveTemplateFS embed.FS

type serveFilters struct {
	Type    string
	Project string
	Tag     string
	Query   string
}

type serveListItem struct {
	ID        string
	Type      string
	Title     string
	Path      string
	TitleHTML template.HTML
	PathHTML  template.HTML
}

type serveIndexData struct {
	Title          string
	CanonicalURL   string
	OGTitle        string
	OGDescription  string
	WorkspaceName  string
	BindInfo       string
	ReadOnly       bool
	GeneratedAt    string
	Filters        serveFilters
	FilterSummary  string
	ActiveChips    []serveFilterChip
	Notes          []serveListItem
	RecentNotes    []serveListItem
	TopTags        []serveTagCount
	Stats          serveStats
	ParseErrors    []string
	HighlightQuery string
	SelectedNoteID string
	Detail         template.HTML
}

type serveNoteDetail struct {
	Note            serveListItem
	Created         string
	Project         string
	Tags            []string
	Frontmatter     []serveKV
	Toc             []serveHeading
	OutboundLinks   []serveOutboundLink
	BrokenLinks     []serveOutboundLink
	BodyHTML        template.HTML
	Backlinks       []serveListItem
	CopyMarkdownURL string
}

type serveStats struct {
	Total     int
	Fleeting  int
	Permanent int
	Project   int
}

type serveTagCount struct {
	Tag   string
	Count int
}

type serveFilterChip struct {
	Key   string
	Value string
	URL   string
}

type serveHeading struct {
	Level int
	Text  string
	ID    string
}

type serveOutboundLink struct {
	Label  string
	Href   string
	ID     string
	Exists bool
}

type serveKV struct {
	Key   string
	Value string
}

type noteServer struct {
	app       *App
	templates *template.Template
	md        goldmark.Markdown
	bindAddr  string
}

var renderedAnchorPattern = regexp.MustCompile(`<a\s+[^>]*href="([^"]+)"[^>]*>`)
var markdownFileIDPattern = regexp.MustCompile(`^(\d{12,14})(?:[-_].*)?\.md$`)
var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func (app *App) runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", "127.0.0.1:8787", "bind address")
	open := fs.Bool("open", false, "open browser after start")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 0 {
		return fmt.Errorf("usage: note serve [--addr <host:port>] [--open]")
	}

	h, err := app.newServeHandler(*addr)
	if err != nil {
		return err
	}

	url := "http://" + *addr
	fmt.Printf("serving notes at %s\n", url)
	if *open {
		_ = openInEditor(url)
	}

	server := &http.Server{
		Addr:    *addr,
		Handler: h,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "received interrupt, shutting down web server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	err = <-errCh
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (app *App) newServeHandler(addr string) (http.Handler, error) {
	tmpl, err := template.ParseFS(serveTemplateFS, "servedata/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse serve templates: %w", err)
	}

	s := &noteServer{
		app:       app,
		templates: tmpl,
		md:        newServeMarkdown(),
		bindAddr:  addr,
	}

	r := chi.NewRouter()
	r.Get("/", s.handleIndex)
	r.Get("/notes", s.handleNotes)
	r.Get("/note/random", s.handleRandom)
	r.Get("/note/{id}", s.handleNote)
	return r, nil
}

func newServeMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithExtensions(
			extension.GFM,
		),
	)
}

func (s *noteServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data, err := s.buildIndexData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data.Title == "" {
		data.Title = data.WorkspaceName + " - note"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *noteServer) handleNotes(w http.ResponseWriter, r *http.Request) {
	data, err := s.buildIndexData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "notes_list", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *noteServer) handleRandom(w http.ResponseWriter, r *http.Request) {
	filters := s.parseFilters(r)
	notes, err := s.app.searchNotes(filters.Type, filters.Project, filters.Tag, filters.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(notes) == 0 {
		http.Error(w, "no notes match current filters", http.StatusNotFound)
		return
	}
	pick := notes[rand.Intn(len(notes))]
	id := pick.Metadata["id"]

	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		detail, err := s.buildNoteDetailData(id, filters.Query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := s.templates.ExecuteTemplate(w, "note_detail", detail); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	q := r.URL.Query()
	q.Set("selected", id)
	http.Redirect(w, r, "/note/"+id+"?"+q.Encode(), http.StatusFound)
}

func (s *noteServer) handleNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		http.Error(w, "missing note id", http.StatusBadRequest)
		return
	}

	filters := s.parseFilters(r)
	detail, err := s.buildNoteDetailData(id, filters.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if strings.EqualFold(r.Header.Get("HX-Request"), "true") {
		if err := s.templates.ExecuteTemplate(w, "note_detail", detail); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	indexData, err := s.buildIndexData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var detailBuf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&detailBuf, "note_detail", detail); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	indexData.Detail = template.HTML(detailBuf.String())
	indexData.SelectedNoteID = id
	indexData.Title = detail.Note.Title + " - note"
	indexData.OGTitle = indexData.Title
	indexData.OGDescription = truncatePlaintext(detail.Note.Title+" "+stripHTML(string(detail.BodyHTML)), 160)
	indexData.CanonicalURL = "/note/" + id

	if err := s.templates.ExecuteTemplate(w, "index", indexData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *noteServer) buildNoteDetailData(id, query string) (serveNoteDetail, error) {
	path, note, _, err := s.app.findNoteByID(id)
	if err != nil {
		return serveNoteDetail{}, err
	}

	var htmlBuf bytes.Buffer
	if err := s.md.Convert([]byte(note.Body), &htmlBuf); err != nil {
		return serveNoteDetail{}, err
	}

	backlinks, err := s.app.backlinks(id)
	if err != nil {
		return serveNoteDetail{}, err
	}

	outbound, broken := s.extractOutboundLinks(note)
	frontmatter := buildFrontmatterKV(note)
	toc := buildTOC(note.Body)

	item := serveListItem{
		ID:    note.Metadata["id"],
		Type:  note.Metadata["type"],
		Title: note.Metadata["title"],
		Path:  relativeToRoot(s.app.root, path),
	}
	if item.Title == "" {
		item.Title = item.ID
	}

	data := serveNoteDetail{
		Note:            item,
		Created:         note.Metadata["created"],
		Project:         note.Metadata["project"],
		Tags:            note.Tags,
		Frontmatter:     frontmatter,
		Toc:             toc,
		OutboundLinks:   outbound,
		BrokenLinks:     broken,
		BodyHTML:        enhanceRenderedLinks(htmlBuf.String(), query),
		Backlinks:       toServeItems(s.app.root, backlinks, query),
		CopyMarkdownURL: fmt.Sprintf("[%s](./%s)", item.Title, filepath.Base(path)),
	}
	return data, nil
}

func (s *noteServer) buildIndexData(r *http.Request) (serveIndexData, error) {
	filters := s.parseFilters(r)
	notes, err := s.app.searchNotes(filters.Type, filters.Project, filters.Tag, filters.Query)
	if err != nil {
		return serveIndexData{}, err
	}

	idx, err := s.app.getIndex()
	if err != nil {
		return serveIndexData{}, err
	}

	stats := serveStats{Total: len(idx.notes)}
	tagCounts := map[string]int{}
	for _, n := range idx.notes {
		switch n.Metadata["type"] {
		case "fleeting":
			stats.Fleeting++
		case "permanent":
			stats.Permanent++
		case "project":
			stats.Project++
		}
		for _, t := range n.Tags {
			if t != "" {
				tagCounts[t]++
			}
		}
	}

	topTags := make([]serveTagCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		topTags = append(topTags, serveTagCount{Tag: tag, Count: count})
	}
	sort.Slice(topTags, func(i, j int) bool {
		if topTags[i].Count != topTags[j].Count {
			return topTags[i].Count > topTags[j].Count
		}
		return topTags[i].Tag < topTags[j].Tag
	})
	if len(topTags) > 12 {
		topTags = topTags[:12]
	}

	recent := idx.notes
	if len(recent) > 8 {
		recent = recent[:8]
	}

	activeChips := buildActiveChips(filters)
	filterSummary := "All notes"
	if len(activeChips) > 0 {
		parts := make([]string, 0, len(activeChips))
		for _, c := range activeChips {
			parts = append(parts, c.Key+": "+c.Value)
		}
		filterSummary = strings.Join(parts, " | ")
	}

	selected := strings.TrimSpace(r.URL.Query().Get("selected"))

	return serveIndexData{
		Title:          filepath.Base(s.app.root) + " - note",
		CanonicalURL:   "/",
		OGTitle:        filepath.Base(s.app.root) + " - note",
		OGDescription:  "Browse markdown notes in a local read-only UI.",
		WorkspaceName:  filepath.Base(s.app.root),
		BindInfo:       s.bindAddr,
		ReadOnly:       true,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Filters:        filters,
		FilterSummary:  filterSummary,
		ActiveChips:    activeChips,
		Notes:          toServeItems(s.app.root, notes, filters.Query),
		RecentNotes:    toServeItems(s.app.root, recent, ""),
		TopTags:        topTags,
		Stats:          stats,
		ParseErrors:    idx.parseErrors,
		HighlightQuery: filters.Query,
		SelectedNoteID: selected,
	}, nil
}

func (s *noteServer) parseFilters(r *http.Request) serveFilters {
	q := r.URL.Query()
	return serveFilters{
		Type:    strings.TrimSpace(q.Get("type")),
		Project: strings.TrimSpace(q.Get("project")),
		Tag:     strings.TrimSpace(q.Get("tag")),
		Query:   strings.TrimSpace(q.Get("q")),
	}
}

func toServeItems(root string, notes []Note, query string) []serveListItem {
	items := make([]serveListItem, 0, len(notes))
	for _, n := range notes {
		title := n.Metadata["title"]
		if title == "" {
			title = n.Metadata["id"]
		}
		relPath := relativeToRoot(root, filepath.Clean(n.Path))
		items = append(items, serveListItem{
			ID:        n.Metadata["id"],
			Type:      n.Metadata["type"],
			Title:     title,
			Path:      relPath,
			TitleHTML: highlightEscaped(title, query),
			PathHTML:  highlightEscaped(relPath, query),
		})
	}
	return items
}

func enhanceRenderedLinks(bodyHTML, query string) template.HTML {
	rewritten := renderedAnchorPattern.ReplaceAllStringFunc(bodyHTML, func(anchorTag string) string {
		match := renderedAnchorPattern.FindStringSubmatch(anchorTag)
		if len(match) < 2 {
			return anchorTag
		}
		href := strings.TrimSpace(match[1])
		id, ok := noteIDFromLinkTarget(href)
		if !ok {
			return anchorTag
		}

		target := "/note/" + id
		updated := strings.Replace(anchorTag, `href="`+href+`"`, `href="`+target+`"`, 1)
		if !strings.Contains(updated, `hx-get=`) {
			updated = strings.TrimSuffix(updated, `>`) +
				` hx-get="` + target + `" hx-target="#note-detail" hx-swap="innerHTML">`
		}
		return updated
	})
	if strings.TrimSpace(query) == "" {
		return template.HTML(rewritten)
	}
	return highlightHTMLText(rewritten, query)
}

func buildActiveChips(filters serveFilters) []serveFilterChip {
	chips := make([]serveFilterChip, 0, 4)
	if filters.Query != "" {
		chips = append(chips, serveFilterChip{Key: "search", Value: filters.Query, URL: buildFilterURL(filters, "q")})
	}
	if filters.Tag != "" {
		chips = append(chips, serveFilterChip{Key: "tag", Value: filters.Tag, URL: buildFilterURL(filters, "tag")})
	}
	if filters.Type != "" {
		chips = append(chips, serveFilterChip{Key: "type", Value: filters.Type, URL: buildFilterURL(filters, "type")})
	}
	if filters.Project != "" {
		chips = append(chips, serveFilterChip{Key: "project", Value: filters.Project, URL: buildFilterURL(filters, "project")})
	}
	return chips
}

func buildFilterURL(filters serveFilters, remove string) string {
	parts := []string{}
	add := func(k, v string) {
		if v != "" {
			parts = append(parts, k+"="+urlEncode(v))
		}
	}
	if remove != "q" {
		add("q", filters.Query)
	}
	if remove != "tag" {
		add("tag", filters.Tag)
	}
	if remove != "type" {
		add("type", filters.Type)
	}
	if remove != "project" {
		add("project", filters.Project)
	}
	if len(parts) == 0 {
		return "/"
	}
	return "/?" + strings.Join(parts, "&")
}

func urlEncode(s string) string {
	repl := strings.NewReplacer(" ", "%20", "#", "%23", "&", "%26", "+", "%2B", "?", "%3F", "=", "%3D")
	return repl.Replace(s)
}

func highlightEscaped(text, query string) template.HTML {
	esc := template.HTMLEscapeString(text)
	q := strings.TrimSpace(query)
	if q == "" {
		return template.HTML(esc)
	}
	lowerEsc := strings.ToLower(esc)
	lowerQ := strings.ToLower(template.HTMLEscapeString(q))
	if lowerQ == "" {
		return template.HTML(esc)
	}
	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lowerEsc[i:], lowerQ)
		if j < 0 {
			b.WriteString(esc[i:])
			break
		}
		j += i
		b.WriteString(esc[i:j])
		b.WriteString("<mark>")
		b.WriteString(esc[j : j+len(lowerQ)])
		b.WriteString("</mark>")
		i = j + len(lowerQ)
	}
	return template.HTML(b.String())
}

func highlightHTMLText(htmlInput, query string) template.HTML {
	q := strings.TrimSpace(query)
	if q == "" {
		return template.HTML(htmlInput)
	}
	parts := htmlTagPattern.Split(htmlInput, -1)
	tags := htmlTagPattern.FindAllString(htmlInput, -1)
	var out strings.Builder
	for i := 0; i < len(parts); i++ {
		out.WriteString(highlightEscapedFragment(parts[i], q))
		if i < len(tags) {
			out.WriteString(tags[i])
		}
	}
	return template.HTML(out.String())
}

// highlightEscapedFragment highlights query matches inside an already-escaped
// HTML text fragment. It must not escape the input again, or entities such as
// &gt; will render literally.
func highlightEscapedFragment(escapedText, query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return escapedText
	}
	lowerText := strings.ToLower(escapedText)
	lowerQ := strings.ToLower(template.HTMLEscapeString(q))
	if lowerQ == "" {
		return escapedText
	}

	var b strings.Builder
	i := 0
	for {
		j := strings.Index(lowerText[i:], lowerQ)
		if j < 0 {
			b.WriteString(escapedText[i:])
			break
		}
		j += i
		b.WriteString(escapedText[i:j])
		b.WriteString("<mark>")
		b.WriteString(escapedText[j : j+len(lowerQ)])
		b.WriteString("</mark>")
		i = j + len(lowerQ)
	}
	return b.String()
}

func buildFrontmatterKV(note Note) []serveKV {
	kvs := make([]serveKV, 0, len(note.Metadata)+1)
	keys := make([]string, 0, len(note.Metadata))
	hasTagsMetadata := false
	for k := range note.Metadata {
		keys = append(keys, k)
		if k == "tags" && strings.TrimSpace(note.Metadata[k]) != "" {
			hasTagsMetadata = true
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if note.Metadata[k] == "" {
			continue
		}
		kvs = append(kvs, serveKV{Key: k, Value: note.Metadata[k]})
	}
	if len(note.Tags) > 0 && !hasTagsMetadata {
		kvs = append(kvs, serveKV{Key: "tags", Value: strings.Join(note.Tags, ", ")})
	}
	return kvs
}

func buildTOC(body string) []serveHeading {
	lines := strings.Split(body, "\n")
	seen := map[string]int{}
	toc := make([]serveHeading, 0)
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "#") {
			continue
		}
		level := 0
		for level < len(trim) && trim[level] == '#' {
			level++
		}
		if level == 0 || level > 6 || len(trim) <= level || trim[level] != ' ' {
			continue
		}
		text := strings.TrimSpace(trim[level:])
		if text == "" {
			continue
		}
		slug := slugifyHeading(text)
		seen[slug]++
		if seen[slug] > 1 {
			slug = fmt.Sprintf("%s-%d", slug, seen[slug])
		}
		toc = append(toc, serveHeading{Level: level, Text: text, ID: slug})
	}
	if len(toc) > 20 {
		return toc[:20]
	}
	return toc
}

func slugifyHeading(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func (s *noteServer) extractOutboundLinks(note Note) ([]serveOutboundLink, []serveOutboundLink) {
	idx, err := s.app.getIndex()
	if err != nil {
		return nil, nil
	}
	idSet := map[string]bool{}
	for _, n := range idx.notes {
		id := n.Metadata["id"]
		if id != "" {
			idSet[id] = true
		}
	}

	seen := map[string]bool{}
	out := make([]serveOutboundLink, 0)
	broken := make([]serveOutboundLink, 0)

	for _, m := range wikiLinkPattern.FindAllStringSubmatch(note.Body, -1) {
		if len(m) < 2 {
			continue
		}
		target := strings.TrimSpace(m[1])
		if target == "" || seen["wiki:"+target] {
			continue
		}
		seen["wiki:"+target] = true
		entry := serveOutboundLink{Label: target, Href: "/note/" + target, ID: target, Exists: idSet[target]}
		out = append(out, entry)
		if !entry.Exists {
			broken = append(broken, entry)
		}
	}

	for _, m := range markdownLinkTargetPattern.FindAllStringSubmatch(note.Body, -1) {
		if len(m) < 2 {
			continue
		}
		raw := strings.TrimSpace(m[1])
		if raw == "" || seen["md:"+raw] {
			continue
		}
		seen["md:"+raw] = true
		entry := serveOutboundLink{Label: raw, Href: raw, Exists: true}
		if id, ok := noteIDFromLinkTarget(raw); ok {
			entry.ID = id
			entry.Href = "/note/" + id
			entry.Exists = idSet[id]
		}
		out = append(out, entry)
		if !entry.Exists {
			broken = append(broken, entry)
		}
	}

	return out, broken
}

func stripHTML(s string) string {
	return strings.TrimSpace(htmlTagPattern.ReplaceAllString(s, " "))
}

func truncatePlaintext(s string, n int) string {
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func noteIDFromLinkTarget(href string) (string, bool) {
	if href == "" || strings.HasPrefix(href, "#") {
		return "", false
	}
	lower := strings.ToLower(href)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:") {
		return "", false
	}

	clean := href
	if i := strings.IndexAny(clean, "?#"); i >= 0 {
		clean = clean[:i]
	}
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "", false
	}

	base := path.Base(clean)
	if m := markdownFileIDPattern.FindStringSubmatch(base); len(m) > 1 {
		return m[1], true
	}
	if digitsOnly(base) && (len(base) == 12 || len(base) == 14) {
		return base, true
	}
	return "", false
}

func digitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
