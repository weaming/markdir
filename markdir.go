package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

var listen = flag.String("listen", "127.0.0.1:10200", "listen host:port")
var showHidden = flag.Bool("all", false, "show hide directories")
var noIndex = flag.String("no-index", "", "comma separated list of directories to disable listing")

func main() {
	flag.Parse()

	httpdir := http.Dir(".")
	handler := renderer{httpdir, http.FileServer(httpdir)}

	log.Printf("Serving on http://%v\n", *listen)
	log.Fatal(http.ListenAndServe(*listen, handler))
}

var outputTemplate = template.Must(template.New("base").Parse(MDTemplate))

type renderer struct {
	d http.Dir
	h http.Handler
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
	".json", ".yaml", ".yml", ".cxx", ".h", ".html", ".ini", ".java", ".js", ".jsp",
	".log", ".map", ".mod", ".sh", ".bash", ".txt", ".xml", ".py", ".go", ".rs",
	".coffee", ".conf", ".config", ".cr", ".css", ".d", ".dart", ".fish", ".gradle",
	".jade", ".json5", ".jsx", ".key", ".less", ".m4", ".markdown", ".md", ".patch",
	".pem", ".plist", ".properties", ".pub", ".pug", ".rb", ".rc", ".sass", ".scpt",
	".scss", ".sql", ".template", ".todo", ".toml", ".ts", ".tsx", ".vim", ".vue",
	".xhtml",
}

func (r renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

		r.h.ServeHTTP(rw, req)

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
		output := blackfriday.Run(input)

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
	r.h.ServeHTTP(rw, req)
}
