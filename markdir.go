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

var codeExtensions = []string{".a", ".asm", ".asp", ".awk", ".bat", ".c", ".class", ".cmd", ".cpp", ".csv", ".json", ".yaml", ".yml", ".cxx", ".h", ".html", ".ini", ".java", ".js", ".jsp", ".log", ".map", ".mod", ".sh", ".bash", ".txt", ".xml", ".py", ".go", ".rs", ".coffee", ".conf", ".config", "cpp", "cr", "css", "d", "dart", "exmaple", "fish", "gradle", "h", "jade", "json5", "jsx", "key", "less", "m4", "markdown", "md", "patch", "pem", "plist", "properties", "pub", "pug", "rb", "rc", "sass", "scpt", "scss", "sql", "template", "todo", "toml", "ts", "tsx", "vim", "vue", "xhtml", "xml"}

func (r renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, ".md") || strings.HasSuffix(req.URL.Path, "/Guide") {
		// net/http is already running a path.Clean on the req.URL.Path,
		// so this is not a directory traversal, at least by my testing
		input, err := ioutil.ReadFile("." + req.URL.Path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", req.URL.Path, err)
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
			Path:         req.URL.Path,
			Body:         template.HTML(string(output)),
			HasCustomCSS: hasCustomCSS,
		})
	} else if hasSuffix(req.URL.Path, codeExtensions) {
		content, err := ioutil.ReadFile("." + req.URL.Path)
		if err != nil {
			http.Error(rw, "not found", 404)
			log.Printf("Couldn't read path %s: %v\n", req.URL.Path, err)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		rw.Write(content)
	} else {
		if isDir(req) {
			name := filepath.Base(req.URL.Path)
			if len(name) >= 2 && name[0] == '.' && name[1] != '.' {
				if !*showHidden {
					http.Error(rw, "not found", 404)
					return
				}
			}
		}
		if isDir(req) {
			out := MDTemplateIndex
			if _, err := os.Stat("index.css"); err == nil {
				out = strings.Replace(out, "</head>", "<link rel=\"stylesheet\" href=\"/index.css\">\n</head>", 1)
			}
			rw.Write([]byte(out))
		}
		r.h.ServeHTTP(rw, req)
		if isDir(req) {
			rw.Write([]byte(MDTemplateIndexTail))
		}
	}
}
