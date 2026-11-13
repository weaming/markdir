package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatFileSize(t *testing.T) {
	cases := []struct {
		size int64
		want string
	}{
		{0, "0.0 B"},
		{1, "1.0 B"},
		{1023, "1023.0 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{2621440, "2.5 MiB"},
		{1 << 30, "1.0 GiB"},
		{1 << 60, "1.0 EiB"},
	}
	for _, test := range cases {
		if got := formatFileSize(test.size); got != test.want {
			t.Errorf("formatFileSize(%d) = %q, want %q", test.size, got, test.want)
		}
	}
}

func testMIMETypes(t *testing.T) map[string]string {
	t.Helper()
	mimeTypes, err := parseMIMETypes(nginxMIMETypes)
	if err != nil {
		t.Fatalf("parse embedded mime.types: %v", err)
	}
	return mimeTypes
}

func TestDirectoryServing(t *testing.T) {
	t.Setenv("TZ", "Asia/Hong_Kong")
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	originalNoIndex := *noIndex
	t.Cleanup(func() {
		*noIndex = originalNoIndex
		if err := os.Chdir(originalDir); err != nil {
			t.Error(err)
		}
	})
	*noIndex = "/site/"

	files := map[string]string{
		"index.html":                   "root home",
		"site/index.html":              "site home",
		"site/style.css":               "body { color: red; }",
		"site/app.js":                  "console.log('ready');",
		"downloads/file.bin":           strings.Repeat("x", 1536),
		"downloads/child/file.bin":     "child",
		"other/file.bin":               "other",
		"listing/index.html/child.txt": "not a homepage",
	}
	for name, content := range files {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	modified := time.Date(2026, time.October, 20, 2, 30, 0, 0, time.UTC)
	if err := os.Chtimes("downloads/file.bin", modified, modified); err != nil {
		t.Fatal(err)
	}
	handler := &renderer{
		handler:        http.FileServer(http.Dir(".")),
		mimeTypes:      testMIMETypes(t),
		indexDirs:      parseIndexDirs(" downloads/ , ,/other/../downloads "),
		gitignoreCache: make(map[string][]gitignoreEntry),
	}
	cases := []struct {
		path     string
		status   int
		contains string
		hasSize  bool
	}{
		{"/", 200, "root home", false},
		{"/site/", 200, "site home", false},
		{"/site", 301, "", false},
		{"/site/index.html", 301, "", false},
		{"/downloads/", 200, "1.5 KiB", true},
		{"/downloads/child/", 200, "file.bin", false},
		{"/other/", 200, "file.bin", false},
		{"/listing/", 200, "index.html/", false},
	}
	for _, test := range cases {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			body := recorder.Body.String()
			if recorder.Code != test.status || !strings.Contains(body, test.contains) {
				t.Fatalf("status = %d, body = %q", recorder.Code, body)
			}
			if got := strings.Contains(body, `<span class="file-size">`); got != test.hasSize {
				t.Errorf("file size present = %v, want %v", got, test.hasSize)
			}
			if test.hasSize {
				want := `file.bin</a><time class="file-modified" datetime="2026-10-20T10:30:00+08:00" title="2026-10-20T10:30:00+08:00">2026-10-20 10:30</time><span class="file-size">1.5 KiB</span>`
				if !strings.Contains(body, want) {
					t.Error("expected modification date and size outside filename link")
				}
			} else if strings.Contains(body, `<time class="file-modified"`) {
				t.Error("unexpected modification date in basic listing")
			}
		})
	}

	for _, asset := range []struct {
		path        string
		contentType string
	}{
		{"/site/", "text/html"},
		{"/site/style.css", "text/css"},
		{"/site/app.js", "javascript"},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, asset.path, nil))
		if got := recorder.Header().Get("Content-Type"); !strings.Contains(got, asset.contentType) {
			t.Errorf("%s Content-Type = %q, want %q", asset.path, got, asset.contentType)
		}
	}

	*noIndex = "/downloads/"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/downloads/", nil))
	if !strings.Contains(recorder.Body.String(), "does not support listing") {
		t.Fatal("directory listing should be disabled")
	}
}

func TestFileContentType(t *testing.T) {
	mimeTypes := testMIMETypes(t)
	dir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Error(err)
		}
	})

	for _, test := range []struct {
		name    string
		content string
		want    string
	}{
		{name: "notes.txt", content: "hello\n", want: "text/plain; charset=utf-8"},
		{name: "data.json", content: "{\"a\": 1}\n", want: "application/json"},
		{name: "app.js", content: "console.log(1)\n", want: "application/javascript"},
		{name: "style.css", content: "body { color: red; }\n", want: "text/css; charset=utf-8"},
		{name: "index.html", content: "<h1>home</h1>\n", want: "text/html; charset=utf-8"},
		{name: "conf.yaml", content: "key: value\n", want: "text/plain; charset=utf-8"},
		{name: "Makefile", content: "all:\n\techo hi\n", want: "text/plain; charset=utf-8"},
		{name: "script.zshrc", content: "alias ll='ls -l'\n", want: "text/plain; charset=utf-8"},
		{name: "main.go", content: "package main\n", want: "text/plain; charset=utf-8"},
		{name: "data.csv", content: "a,b\n1,2\n", want: "text/plain; charset=utf-8"},
		{name: "app.log", content: "started\n", want: "text/plain; charset=utf-8"},
		{name: "archive.zip", content: "PK\x03\x04binary", want: "application/zip"},
		{name: "photo.png", content: "\x89PNG\r\n\x1a\n", want: "image/png"},
		{name: "font.woff2", content: "wOF2\x00\x00", want: "font/woff2"},
		{name: "blob.bin", content: "\x00\x01\x02\x03", want: "application/octet-stream"},
		{name: "blob.unknownext", content: "plain text here\n", want: "text/plain; charset=utf-8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(test.name, []byte(test.content), 0644); err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			setFileContentType(recorder, "/"+test.name, mimeTypes)
			if got := recorder.Header().Get("Content-Type"); got != test.want {
				t.Fatalf("%s Content-Type = %q, want %q", test.name, got, test.want)
			}
		})
	}
}

func TestListingTimezone(t *testing.T) {
	for _, test := range []struct {
		zone   string
		offset int
	}{
		{"", 8 * 60 * 60},
		{"UTC", 0},
		{"America/New_York", -4 * 60 * 60},
		{"invalid/timezone", 8 * 60 * 60},
	} {
		t.Run(test.zone, func(t *testing.T) {
			t.Setenv("TZ", test.zone)
			_, offset := time.Date(2026, time.October, 20, 0, 0, 0, 0, listingTimezone()).Zone()
			if offset != test.offset {
				t.Errorf("offset = %d, want %d", offset, test.offset)
			}
		})
	}
}
