package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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
var ignore = flag.String("ignore", "", "comma separated list of directories to hide from listing (default: .git/)")
var reverseSort = flag.Bool("reverse", false, "reverse file name sort order")
var hideIcon = flag.Bool("hide-icon", false, "hide icon image files (e.g. icon.png, icon.jpg) from directory listing")
var tocFile = flag.String("toc", "", "path to JSON file mapping URL paths to friendly display names")
var columns = flag.Int("columns", 1, "number of columns for directory listing on desktop")
var exts = flag.String("exts", "", "comma separated extensions to show in listing, e.g. 'json,md,jsonl' (dirs always shown)")
var llms = flag.Bool("llms", false, "llms.txt style, see https://llmstxt.org")

func generateTOC(path string, extFilter []string) map[string]string {
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
		if len(extFilter) > 0 && !dirContainsExts(urlPath, extFilter) {
			return nil
		}
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

func loadTOC(path string, extFilter []string) map[string]string {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return generateTOC(path, extFilter)
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
		dir:            httpdir,
		handler:        http.FileServer(httpdir),
		reverse:        *reverseSort,
		hideIcon:       *hideIcon,
		columns:        *columns,
		exts:           parseExts(*exts),
		ignore:         parseIgnore(*ignore),
		llms:           *llms,
		toc:            loadTOC(*tocFile, parseExts(*exts)),
		gitignoreCache: make(map[string][]gitignoreEntry),
		mdTitleCache:   make(map[string]mdTitleEntry),
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

type mdTitleEntry struct {
	title    string
	cachedAt time.Time
}

type renderer struct {
	dir            http.Dir
	handler        http.Handler
	reverse        bool
	hideIcon       bool
	columns        int
	exts           []string
	ignore         []string
	llms           bool
	tocMu          sync.RWMutex
	toc            map[string]string
	gitignoreCache map[string][]gitignoreEntry
	gitignoreMu    sync.RWMutex
	mdTitleCache   map[string]mdTitleEntry
	mdTitleMu      sync.RWMutex
}

func (r *renderer) getTOC(key string) (string, bool) {
	r.tocMu.RLock()
	defer r.tocMu.RUnlock()
	v, ok := r.toc[key]
	return v, ok
}

func (r *renderer) getGitignore(dirPath string) []gitignoreEntry {
	r.gitignoreMu.RLock()
	entries, ok := r.gitignoreCache[dirPath]
	r.gitignoreMu.RUnlock()
	if ok {
		return entries
	}

	var allEntries []gitignoreEntry
	for p := dirPath; ; {
		gitignorePath := filepath.Join(p, ".gitignore")
		if parsed, err := parseGitignore(gitignorePath); err == nil {
			allEntries = append(allEntries, parsed...)
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}

	r.gitignoreMu.Lock()
	r.gitignoreCache[dirPath] = allEntries
	r.gitignoreMu.Unlock()
	return allEntries
}

func (r *renderer) isGitignored(name string, isDir bool, dirPath string) bool {
	entries := r.getGitignore(dirPath)
	if len(entries) == 0 {
		return false
	}
	return matchGitignore(name, isDir, entries)
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

func dirContainsExts(urlPath string, exts []string) bool {
	fsPath := "." + urlPath
	found := false
	filepath.WalkDir(fsPath, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if hasSuffix(strings.ToLower(d.Name()), exts) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func parseIgnore(raw string) []string {
	defaultIgnore := []string{".git/"}
	if raw == "" {
		return defaultIgnore
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts)+len(defaultIgnore))
	result = append(result, defaultIgnore...)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		result = append(result, p)
	}
	return result
}

func isIgnored(name string, ignoreList []string) bool {
	for _, ign := range ignoreList {
		if name == ign || strings.HasPrefix(name, ign) || strings.HasPrefix(ign, name) {
			return true
		}
	}
	return false
}

// gitignoreEntry represents a parsed gitignore pattern.
type gitignoreEntry struct {
	pattern string
	negate  bool
	dirOnly bool
}

// parseGitignore reads a .gitignore file and returns patterns.
func parseGitignore(path string) ([]gitignoreEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []gitignoreEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, " \t\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		negate := false
		if strings.HasPrefix(line, "!") {
			negate = true
			line = line[1:]
		}
		dirOnly := strings.HasSuffix(line, "/")
		if dirOnly {
			line = line[:len(line)-1]
		}
		if line != "" {
			entries = append(entries, gitignoreEntry{
				pattern: line,
				negate:  negate,
				dirOnly: dirOnly,
			})
		}
	}
	return entries, nil
}

// matchGitignore checks if name matches any gitignore entry.
func matchGitignore(name string, isDir bool, entries []gitignoreEntry) bool {
	matched := false
	nameWithSlash := name + "/"
	for _, e := range entries {
		if e.dirOnly && !isDir {
			continue
		}
		hit := matchGitignorePattern(name, e.pattern)
		if !hit && isDir {
			hit = matchGitignorePattern(nameWithSlash, e.pattern)
		}
		if hit {
			matched = !e.negate
		}
	}
	return matched
}

func matchGitignorePattern(name, pattern string) bool {
	if pattern == "" {
		return false
	}
	if suffix, ok := strings.CutPrefix(pattern, "**/"); ok {
		// **/foo matches foo anywhere in the path
		return name == suffix || strings.HasSuffix(name, "/"+suffix)
	}
	if prefix, ok := strings.CutSuffix(pattern, "/**"); ok {
		return name == prefix || strings.HasPrefix(name, prefix+"/")
	}
	if strings.Contains(pattern, "/") {
		matched, _ := filepath.Match(pattern, name)
		return matched
	}
	// No slash: match against each path component
	for _, part := range strings.Split(name, "/") {
		if matched, _ := filepath.Match(pattern, part); matched {
			return true
		}
	}
	return false
}

func parseExts(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		result = append(result, strings.ToLower(p))
	}
	return result
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

// dirTitle derives a page title from a URL path segment:
// replaces '_' and '-' with spaces, then title-cases each word.
func dirTitle(urlPath string) string {
	name := strings.TrimSuffix(urlPath, "/")
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if name == "" {
		wd, _ := os.Getwd()
		name = filepath.Base(wd)
	}
	name = strings.NewReplacer("_", " ", "-", " ").Replace(name)
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func indexHead(urlPath string) string {
	head := strings.Replace(MDTemplateIndex, "{{TITLE}}", dirTitle(urlPath), 1)
	if _, err := os.Stat("index.css"); err == nil {
		head = strings.Replace(head, "</head>", "<link rel=\"stylesheet\" href=\"/index.css\">\n</head>", 1)
	}
	return head
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

	// 1. Handle /llms.txt virtual endpoint
	if r.llms && path == "/llms.txt" {
		log.Println(path)
		r.serveLLMSTxt(rw, req)
		return
	}

	// 2. Handle /date/ virtual paths
	if path == "/date" {
		http.Redirect(rw, req, "/date/", http.StatusMovedPermanently)
		return
	}
	if path == "/date/" {
		log.Println(path)
		r.serveDateList(rw, req)
		return
	}
	if strings.HasPrefix(path, "/date/") && isDir(req) {
		rest := strings.TrimPrefix(path, "/date/")
		if date, remainder, ok := strings.Cut(rest, "/"); ok {
			subpath := strings.TrimSuffix(remainder, "/")
			if isDateString(date) {
				log.Println(path)
				r.serveDateDirectoryListing(rw, req, date, subpath)
				return
			}
		}
	}

	// 2. Check for directory
	if isDir(req) {
		log.Println(path)
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
		rw.Write([]byte(indexHead(req.URL.Path)))

		if r.reverse {
			r.serveDirectoryListing(rw, req, true)
		} else {
			r.serveDirectoryListing(rw, req, false)
		}

		rw.Write([]byte(MDTemplateIndexTail))
		return
	}

	// 2. Handle /date/{date}/{source path} → /{source path stem}/{date}{ext}
	if strings.HasPrefix(path, "/date/") && !isDir(req) {
		rest := strings.TrimPrefix(path, "/date/")
		if date, sourcePath, ok := strings.Cut(rest, "/"); ok {
			ext := filepath.Ext(sourcePath)
			if isDateString(date) && ext != "" {
				sourceWithoutExt := strings.TrimSuffix(sourcePath, ext)
				remappedPath := "/" + sourceWithoutExt + "/" + date + ext
				log.Printf("%s -> .%s", path, remappedPath)
				newReq := *req
				newURL := *req.URL
				newURL.Path = remappedPath
				newReq.URL = &newURL
				r.ServeHTTP(rw, &newReq)
				return
			}
			// llms: extensionless date path → remap without .md so stripped-URL handler renders HTML
			if r.llms && isDateString(date) && ext == "" {
				remappedPath := "/" + sourcePath + "/" + date
				log.Printf("%s -> .%s (llms)", path, remappedPath)
				newReq := *req
				newURL := *req.URL
				newURL.Path = remappedPath
				newReq.URL = &newURL
				r.ServeHTTP(rw, &newReq)
				return
			}
		}
	}

	// 3. llms mode: .md URL → raw plaintext markdown
	if r.llms && strings.HasSuffix(path, ".md") {
		log.Println(path)
		input, err := os.ReadFile("." + path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", path, err)
			return
		}
		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.Write(input)
		return
	}

	// 4. Handle Markdown files → render HTML.
	// In llms mode, listing links have .md stripped, so we resolve path+".md".
	mdFilePath := "." + path
	isMdURL := strings.HasSuffix(path, ".md") || strings.HasSuffix(path, "/Guide")
	if !isMdURL && r.llms && !isDir(req) {
		candidate := "." + path + ".md"
		if _, err := os.Stat(candidate); err == nil {
			mdFilePath = candidate
			isMdURL = true
		}
	}
	if isMdURL {
		log.Println(path)
		input, err := os.ReadFile(mdFilePath)
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

		hasCustomCSS := false
		if _, err := os.Stat("index.css"); err == nil {
			hasCustomCSS = true
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		outputTemplate.Execute(rw, struct {
			Path         string
			Body         template.HTML
			HasCustomCSS bool
		}{
			Path:         path,
			Body:         template.HTML(buf.String()),
			HasCustomCSS: hasCustomCSS,
		})
		return
	}

	// 5. Handle Code files
	if hasSuffix(path, codeExtensions) {
		log.Println(path)
		content, err := os.ReadFile("." + path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", path, err)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.Write(content)
		return
	}

	// 6. Default: serve as is
	r.handler.ServeHTTP(rw, req)
}

func (r *renderer) serveDirectoryListing(rw http.ResponseWriter, req *http.Request, reverse bool) {
	fullPath := "." + req.URL.Path
	entries, err := os.ReadDir(fullPath)
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
		displayName := name
		isDir := entry.IsDir()
		if isDir {
			displayName += "/"
		}
		if isIgnored(displayName, r.ignore) {
			continue
		}
		if r.isGitignored(name, isDir, fullPath) {
			continue
		}
		if r.hideIcon && isIconFile(name) {
			continue
		}
		if len(r.exts) > 0 {
			if isDir {
				if !dirContainsExts(req.URL.Path+name+"/", r.exts) {
					continue
				}
			} else if !hasSuffix(strings.ToLower(name), r.exts) {
				continue
			}
		}
		if isDir {
			name += "/"
		}
		urlPath := req.URL.Path + name
		if r.llms && strings.HasSuffix(urlPath, ".md") {
			urlPath = strings.TrimSuffix(urlPath, ".md")
			displayName = strings.TrimSuffix(displayName, ".md")
		}
		if label, ok := r.getTOC(urlPath); ok {
			displayName = label
		}
		fmt.Fprintf(rw, "<li><a href=\"%s\">%s</a></li>\n", urlPath, displayName)
	}
	rw.Write([]byte("</ul>\n"))
}

// serveDateDirectoryListing 渲染 /date/{date}/{subpath}/ 虚拟目录。
// subpath 为空时扫描根目录，否则扫描 ./{subpath}/。
// 若子目录直接含 {date}.md，链接到 .md 页；否则若递归包含则链接到子目录继续浏览。
func (r *renderer) serveDateDirectoryListing(rw http.ResponseWriter, req *http.Request, date, subpath string) {
	scanDir := "."
	if subpath != "" {
		scanDir = "./" + subpath
	}

	// 单次 Glob 获取 scanDir 下所有子目录中匹配 date 的文件
	globMatches, _ := filepath.Glob(filepath.Join(scanDir, "*", date+".*"))
	directMap := map[string][]string{}
	for _, m := range globMatches {
		filename := filepath.Base(m)
		if len(r.exts) > 0 && !hasSuffix(strings.ToLower(filename), r.exts) {
			continue
		}
		subdir := filepath.Base(filepath.Dir(m))
		if strings.HasPrefix(subdir, ".") {
			continue
		}
		directMap[subdir] = append(directMap[subdir], filename)
	}

	// 枚举子目录，找出无直接文件但有深层内容的（用于导航链接）
	allEntries, err := os.ReadDir(scanDir)
	if err != nil {
		http.Error(rw, "cannot read directory", http.StatusInternalServerError)
		return
	}
	navDirs := map[string]struct{}{}
	for _, entry := range allEntries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		name := entry.Name()
		if _, ok := directMap[name]; ok {
			continue
		}
		// 用特定日期 Glob 再深一层，避免用近 N 天误导用户
		deeper, _ := filepath.Glob(filepath.Join(scanDir, name, "*", date+".*"))
		for _, m := range deeper {
			if len(r.exts) == 0 || hasSuffix(strings.ToLower(filepath.Base(m)), r.exts) {
				navDirs[name] = struct{}{}
				break
			}
		}
	}

	// 合并所有条目并排序
	allNames := make([]string, 0, len(directMap)+len(navDirs))
	for name := range directMap {
		allNames = append(allNames, name)
	}
	for name := range navDirs {
		allNames = append(allNames, name)
	}
	if r.reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(allNames)))
	} else {
		sort.Strings(allNames)
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(indexHead(req.URL.Path)))

	if r.columns > 1 {
		fmt.Fprintf(rw, "<style>@media(min-width:601px){.dir-list{display:grid;grid-template-columns:repeat(%d,1fr);}}</style>\n", r.columns)
	}
	rw.Write([]byte("<ul class=\"dir-list\">\n"))

	for _, name := range allNames {
		tocKey := "/" + name + "/"
		if subpath != "" {
			tocKey = "/" + subpath + "/" + name + "/"
		}
		displayName := name
		if label, ok := r.getTOC(tocKey); ok {
			displayName = label
		}

		if files, ok := directMap[name]; ok {
			for _, filename := range files {
				ext := filepath.Ext(filename)
				urlPath := req.URL.Path + name + ext
				if r.llms && ext == ".md" {
					urlPath = strings.TrimSuffix(urlPath, ".md")
				}
				fmt.Fprintf(rw, "<li><a href=\"%s\">%s</a></li>\n", urlPath, displayName)
			}
		} else {
			fmt.Fprintf(rw, "<li><a href=\"%s\">%s</a></li>\n", req.URL.Path+name+"/", displayName)
		}
	}

	rw.Write([]byte("</ul>\n"))
	rw.Write([]byte(MDTemplateIndexTail))
}

// serveDateList 渲染 /date/ 虚拟页，列出所有 Y-m-d 格式的日期。
// 用单次 Glob 扫描顶层子目录下的所有文件，收集唯一日期，按倒序排列。
func (r *renderer) serveDateList(rw http.ResponseWriter, req *http.Request) {
	globMatches, err := filepath.Glob(filepath.Join(".", "*", "*"))
	if err != nil {
		http.Error(rw, "cannot read directory", http.StatusInternalServerError)
		return
	}

	dateSet := map[string]struct{}{}
	for _, m := range globMatches {
		name := filepath.Base(m)
		ext := filepath.Ext(name)
		if ext == "" {
			continue // 无扩展名视为目录，跳过
		}
		// 跳过以 . 开头的顶层目录下的文件
		parts := strings.SplitN(filepath.ToSlash(m), "/", 3)
		if len(parts) >= 2 && strings.HasPrefix(parts[1], ".") {
			continue
		}
		stem := strings.TrimSuffix(name, ext)
		if len(r.exts) > 0 && !hasSuffix(strings.ToLower(name), r.exts) {
			continue
		}
		if isDateString(stem) {
			dateSet[stem] = struct{}{}
		}
	}

	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(indexHead(req.URL.Path)))
	if r.columns > 1 {
		fmt.Fprintf(rw, "<style>@media(min-width:601px){.dir-list{display:grid;grid-template-columns:repeat(%d,1fr);}}</style>\n", r.columns)
	}
	rw.Write([]byte("<ul class=\"dir-list\">\n"))
	for _, date := range dates {
		fmt.Fprintf(rw, "<li><a href=\"/date/%s/\">%s</a></li>\n", date, date)
	}
	rw.Write([]byte("</ul>\n"))
	rw.Write([]byte(MDTemplateIndexTail))
}

// isDateString 检查字符串是否为 YYYY-MM-DD 格式。
func isDateString(s string) bool {
	if len(s) != 10 {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// readMDTitle scans a markdown file for the first H1 line and returns its text.
func readMDTitle(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}

func (r *renderer) getMDTitle(filePath string) string {
	r.mdTitleMu.RLock()
	entry, ok := r.mdTitleCache[filePath]
	r.mdTitleMu.RUnlock()
	if ok && time.Since(entry.cachedAt) < time.Minute {
		return entry.title
	}
	title := readMDTitle(filePath)
	r.mdTitleMu.Lock()
	r.mdTitleCache[filePath] = mdTitleEntry{title: title, cachedAt: time.Now()}
	r.mdTitleMu.Unlock()
	return title
}

// serveLLMSTxt generates /llms.txt following the llmstxt.org spec:
// H1 title, then H2 sections per top-level directory, each listing .md files.
func (r *renderer) serveLLMSTxt(rw http.ResponseWriter, req *http.Request) {
	// Collect root-level .md files.
	var rootFiles []string
	// Enumerate top-level subdirectories in order.
	topEntries, _ := os.ReadDir(".")
	var sections []string
	for _, e := range topEntries {
		if !e.IsDir() {
			if strings.HasSuffix(e.Name(), ".md") {
				if len(r.exts) == 0 || hasSuffix(strings.ToLower(e.Name()), r.exts) {
					rootFiles = append(rootFiles, "/"+e.Name())
				}
			}
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		sections = append(sections, e.Name())
	}

	// For each section, find all .md files recursively.
	groups := map[string][]string{}
	for _, section := range sections {
		filepath.WalkDir("./"+section, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name() != section && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			if len(r.exts) > 0 && !hasSuffix(strings.ToLower(d.Name()), r.exts) {
				return nil
			}
			groups[section] = append(groups[section], "/"+strings.TrimPrefix(filepath.ToSlash(p), "./"))
			return nil
		})
		if r.reverse {
			sort.Sort(sort.Reverse(sort.StringSlice(groups[section])))
		} else {
			sort.Strings(groups[section])
		}
	}

	wd, _ := os.Getwd()
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n", dirTitle("/"+filepath.Base(wd)+"/"))

	if len(rootFiles) > 0 {
		sb.WriteString("\n")
		for _, filePath := range rootFiles {
			title := r.getMDTitle("." + filePath)
			if title == "" {
				title = strings.TrimSuffix(filepath.Base(filePath), ".md")
			}
			fmt.Fprintf(&sb, "- [%s](%s)\n", title, filePath)
		}
	}

	for _, section := range sections {
		if len(groups[section]) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "\n## %s\n\n", dirTitle("/"+section+"/"))
		for _, filePath := range groups[section] {
			title := r.getMDTitle("." + filePath)
			if title == "" {
				title = strings.TrimSuffix(filepath.Base(filePath), ".md")
			}
			fmt.Fprintf(&sb, "- [%s](%s)\n", title, filePath)
		}
	}

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.Write([]byte(sb.String()))
}
