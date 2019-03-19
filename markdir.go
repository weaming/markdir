package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/russross/blackfriday"
)

var bind = flag.String("bind", "127.0.0.1:10200", "listen host:port")

func main() {
	flag.Parse()

	httpdir := http.Dir(".")
	handler := renderer{httpdir, http.FileServer(httpdir)}

	fmt.Printf("Serving on http://%v\n", *bind)
	log.Fatal(http.ListenAndServe(*bind, handler))
}

var outputTemplate = template.Must(template.New("base").Parse(MDTemplate))

type renderer struct {
	d http.Dir
	h http.Handler
}

func isDir(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, "/")
}

func (r renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !strings.HasSuffix(req.URL.Path, ".md") {
		if isDir(req) {
			rw.Write([]byte(MDTemplateIndex))
		}
		r.h.ServeHTTP(rw, req)
		if isDir(req) {
			rw.Write([]byte(MDTemplateIndexTail))
		}
		return
	}

	// net/http is already running a path.Clean on the req.URL.Path,
	// so this is not a directory traversal, at least by my testing
	input, err := ioutil.ReadFile("." + req.URL.Path)
	if err != nil {
		http.Error(rw, "Internal Server Error", 500)
		log.Fatalf("Couldn't read path %s: %v", req.URL.Path, err)
	}
	output := blackfriday.Run(input)

	rw.Header().Set("Content-Type", "text/html")

	outputTemplate.Execute(rw, struct {
		Path string
		Body template.HTML
	}{
		Path: req.URL.Path,
		Body: template.HTML(string(output)),
	})

}
