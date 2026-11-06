package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 内置前端依赖，避免页面依赖外部 CDN。
// 升级方式：
//
//	curl -L -o assets/mermaid.min.js https://cdn.jsdelivr.net/npm/mermaid@<version>/dist/mermaid.min.js
//	curl -L -o assets/html-to-image.js https://cdn.jsdelivr.net/npm/html-to-image@<version>/dist/html-to-image.js
//
//go:embed assets/mermaid.min.js
var mermaidJS []byte

//go:embed assets/html-to-image.js
var htmlToImageJS []byte

// 资源带版本号查询串，升级文件时同步改这里，避免 immutable 缓存长期吃旧版
const (
	mermaidVersion     = "11.17.2"
	htmlToImageVersion = "1.11.11"
)

var (
	gzipOnce    sync.Once
	mermaidGzip []byte
	htmlToImgGz []byte
)

// compressAssets 首次请求时压缩一次，之后复用同一份字节
func compressAssets() {
	gzipOnce.Do(func() {
		mermaidGzip = gzipBytes(mermaidJS)
		htmlToImgGz = gzipBytes(htmlToImageJS)
	})
}

func gzipBytes(source []byte) []byte {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(source); err != nil {
		log.Printf("compress asset: %v", err)
		return nil
	}
	if err := writer.Close(); err != nil {
		log.Printf("compress asset: %v", err)
		return nil
	}
	return buf.Bytes()
}

// assetBody 返回资源原始字节和 gzip 字节
func assetBody(path string) (raw, compressed []byte, ok bool) {
	compressAssets()
	switch path {
	case "/-/mermaid.min.js":
		return mermaidJS, mermaidGzip, true
	case "/-/html-to-image.js":
		return htmlToImageJS, htmlToImgGz, true
	}
	return nil, nil, false
}

// serveAsset 提供内嵌资源，路径以 /-/ 开头，优先于文档目录
func serveAsset(rw http.ResponseWriter, req *http.Request) {
	raw, compressed, ok := assetBody(req.URL.Path)
	if !ok {
		http.NotFound(rw, req)
		return
	}

	body := raw
	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && compressed != nil {
		rw.Header().Set("Content-Encoding", "gzip")
		body = compressed
	}

	rw.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	rw.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	rw.Header().Set("Vary", "Accept-Encoding")
	http.ServeContent(rw, req, filepath.Base(req.URL.Path), time.Time{}, bytes.NewReader(body))
}
