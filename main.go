package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var listen = flag.String("listen", "127.0.0.1:10200", "listen host:port")
var showHidden = flag.Bool("all", false, "show hide directories")
var noIndex = flag.String("no-index", "", "comma separated list of directories to disable listing")
var reverseSort = flag.Bool("reverse", false, "reverse file name sort order")
var hideIcon = flag.Bool("hide-icon", false, "hide icon image files (e.g. icon.png, icon.jpg) from directory listing")
var tocFile = flag.String("toc", "", "path to JSON file mapping URL paths to friendly display names")
var columns = flag.Int("columns", 1, "number of columns for directory listing on desktop")

func generateTOC(path string) map[string]string {
	meta := map[string]string{}

	err := filepath.WalkDir(".", func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == "." {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		depth := strings.Count(filepath.ToSlash(p), "/") + 1
		if depth > 2 {
			return filepath.SkipDir
		}
		urlPath := "/" + filepath.ToSlash(p) + "/"
		meta[urlPath] = d.Name() + "/"
		return nil
	})
	if err != nil {
		log.Fatalf("cannot scan directories for toc: %v", err)
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		log.Fatalf("cannot marshal toc: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("cannot write toc file %s: %v", path, err)
	}
	log.Printf("Generated toc file: %s", path)
	return meta
}

func loadTOC(path string) map[string]string {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return generateTOC(path)
	}
	if err != nil {
		log.Fatalf("cannot read toc file %s: %v", path, err)
	}
	var toc map[string]string
	if err := json.Unmarshal(data, &toc); err != nil {
		log.Fatalf("cannot parse toc file %s: %v", path, err)
	}
	return toc
}

func main() {
	flag.Parse()

	httpdir := http.Dir(".")
	handler := &renderer{
		dir:      httpdir,
		handler:  http.FileServer(httpdir),
		reverse:  *reverseSort,
		hideIcon: *hideIcon,
		columns:  *columns,
		toc:      loadTOC(*tocFile),
	}
	if *tocFile != "" {
		go handler.watchTOC(*tocFile)
	}

	log.Printf("Serving on http://%v\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, handler))
}

var outputTemplate = template.Must(template.New("base").Parse(MDTemplate))

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,         // 表格、删除线、任务列表、自动链接
		extension.Typographer, // 智能标点
		extension.Linkify,     // URL 自动链接
		extension.NewCJK(),    // CJK 换行优化
	),
)

var iconImageExtensions = []string{
	".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".svg",
}

type renderer struct {
	dir      http.Dir
	handler  http.Handler
	reverse  bool
	hideIcon bool
	columns  int
	tocMu    sync.RWMutex
	toc      map[string]string
}

func (r *renderer) getTOC(key string) (string, bool) {
	r.tocMu.RLock()
	defer r.tocMu.RUnlock()
	v, ok := r.toc[key]
	return v, ok
}

func (r *renderer) watchTOC(path string) {
	var lastMod time.Time
	if info, err := os.Stat(path); err == nil {
		lastMod = info.ModTime()
	}

	for range time.Tick(2 * time.Second) {
		info, err := os.Stat(path)
		if err != nil || !info.ModTime().After(lastMod) {
			continue
		}
		lastMod = info.ModTime()

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("toc reload: cannot read %s: %v", path, err)
			continue
		}
		var toc map[string]string
		if err := json.Unmarshal(data, &toc); err != nil {
			log.Printf("toc reload: cannot parse %s: %v", path, err)
			continue
		}
		r.tocMu.Lock()
		r.toc = toc
		r.tocMu.Unlock()
		log.Printf("toc reloaded: %s", path)
	}
}

func isIconFile(name string) bool {
	lname := strings.ToLower(name)
	ext := filepath.Ext(lname)
	base := strings.TrimSuffix(lname, ext)
	if base != "icon" {
		return false
	}
	return hasSuffix(lname, iconImageExtensions)
}

func isDir(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, "/")
}

func hasSuffix(text string, list []string) bool {
	for _, s := range list {
		if strings.HasSuffix(text, s) {
			return true
		}
	}
	return false
}

func isNoIndex(path string) bool {
	if *noIndex == "" {
		return false
	}
	parts := strings.Split(*noIndex, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == path {
			return true
		}
	}
	return false
}

var codeExtensions = []string{
	".a", ".asm", ".asp", ".awk", ".bat", ".c", ".class", ".cmd", ".cpp", ".csv",
	".json", ".jsonl", ".yaml", ".yml", ".cxx", ".h", ".html", ".ini", ".java", ".js", ".jsp",
	".log", ".map", ".mod", ".sh", ".bash", ".txt", ".xml", ".py", ".go", ".rs",
	".coffee", ".conf", ".config", ".cr", ".css", ".d", ".dart", ".fish", ".gradle",
	".jade", ".json5", ".jsx", ".key", ".less", ".m4", ".markdown", ".md", ".patch",
	".pem", ".plist", ".properties", ".pub", ".pug", ".rb", ".rc", ".sass", ".scpt",
	".scss", ".sql", ".template", ".todo", ".toml", ".ts", ".tsx", ".vim", ".vue",
	".xhtml",
}

func (r *renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// 1. Check for directory
	if isDir(req) {
		if isNoIndex(path) {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.Write([]byte("<h3>Current directory does not support listing</h3>"))
			return
		}

		name := filepath.Base(path)
		if len(name) >= 2 && name[0] == '.' && name[1] != '.' {
			if !*showHidden {
				http.Error(rw, "not found", 404)
				return
			}
		}

		// Pre-render directory listing parts
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		outHead := MDTemplateIndex
		if _, err := os.Stat("index.css"); err == nil {
			outHead = strings.Replace(outHead, "</head>", "<link rel=\"stylesheet\" href=\"/index.css\">\n</head>", 1)
		}
		rw.Write([]byte(outHead))

		if r.reverse {
			r.serveDirectoryListing(rw, req, true)
		} else {
			r.serveDirectoryListing(rw, req, false)
		}

		rw.Write([]byte(MDTemplateIndexTail))
		return
	}

	// 2. Handle Markdown files
	if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, "/Guide") {
		input, err := ioutil.ReadFile("." + path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", path, err)
			return
		}
		var buf bytes.Buffer
		if err := md.Convert(input, &buf); err != nil {
			http.Error(rw, "markdown parsing failed", 500)
			log.Printf("Couldn't parse markdown %s: %v\n", path, err)
			return
		}
		output := buf.Bytes()

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")

		hasCustomCSS := false
		if _, err := os.Stat("index.css"); err == nil {
			hasCustomCSS = true
		}
		outputTemplate.Execute(rw, struct {
			Path         string
			Body         template.HTML
			HasCustomCSS bool
		}{
			Path:         path,
			Body:         template.HTML(string(output)),
			HasCustomCSS: hasCustomCSS,
		})
		return
	}

	// 3. Handle Code files
	if hasSuffix(path, codeExtensions) {
		content, err := ioutil.ReadFile("." + path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", path, err)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.Write(content)
		return
	}

	// 4. Default: serve as is
	r.handler.ServeHTTP(rw, req)
}

func (r *renderer) serveDirectoryListing(rw http.ResponseWriter, req *http.Request, reverse bool) {
	fullPath := "." + req.URL.Path
	entries, err := ioutil.ReadDir(fullPath)
	if err != nil {
		http.Error(rw, "cannot read directory", http.StatusInternalServerError)
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		if reverse {
			return entries[i].Name() > entries[j].Name()
		}
		return entries[i].Name() < entries[j].Name()
	})

	if r.columns > 1 {
		fmt.Fprintf(rw, "<style>@media(min-width:601px){.dir-list{display:grid;grid-template-columns:repeat(%d,1fr);}}</style>\n", r.columns)
	}
	rw.Write([]byte("<ul class=\"dir-list\">\n"))
	for _, entry := range entries {
		name := entry.Name()
		if r.hideIcon && isIconFile(name) {
			continue
		}
		if entry.IsDir() {
			name += "/"
		}
		urlPath := req.URL.Path + name
		displayName := name
		if label, ok := r.getTOC(urlPath); ok {
			displayName = label
		}
		fmt.Fprintf(rw, "<li><a href=\"%s\">%s</a></li>\n", urlPath, displayName)
	}
	rw.Write([]byte("</ul>\n"))
}
