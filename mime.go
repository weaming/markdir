package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed mime.types
var nginxMIMETypes string

// renderedTypes 是浏览器本来就会内联渲染、必须保留原类型的内容类型
// （否则页面会被当成源码展示）。
var renderedTypes = []string{
	"text/", "image/", "video/", "audio/", "font/",
	"application/javascript", "application/ecmascript",
	"application/json", "application/xhtml+xml", "application/xml",
}

// textSniffLen 与 http.DetectContentType 的嗅探窗口一致。
const textSniffLen = 512

var (
	textExtCacheMu sync.RWMutex
	textExtCache   = make(map[string]bool)
)

func parseMIMETypes(source string) (map[string]string, error) {
	var uncommented strings.Builder
	for _, line := range strings.Split(source, "\n") {
		line, _, _ = strings.Cut(line, "#")
		uncommented.WriteString(line)
		uncommented.WriteByte('\n')
	}

	tokenized := strings.NewReplacer("{", " { ", "}", " } ", ";", " ; ").Replace(uncommented.String())
	tokens := strings.Fields(tokenized)
	blockStart := -1
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i] == "types" && tokens[i+1] == "{" {
			blockStart = i + 2
			break
		}
	}
	if blockStart < 0 {
		return nil, fmt.Errorf("MIME types block not found")
	}

	mimeTypes := make(map[string]string)
	contentType := ""
	hasBlockEnd := false
	for _, token := range tokens[blockStart:] {
		switch token {
		case "}":
			if contentType != "" {
				return nil, fmt.Errorf("MIME type %q is missing a semicolon", contentType)
			}
			hasBlockEnd = true
			break
		case ";":
			if contentType == "" {
				return nil, fmt.Errorf("unexpected semicolon in MIME types block")
			}
			contentType = ""
		default:
			if contentType == "" {
				contentType = token
				continue
			}
			ext := strings.ToLower(strings.TrimPrefix(token, "."))
			if ext == "" {
				return nil, fmt.Errorf("empty extension for MIME type %q", contentType)
			}
			mimeTypes[ext] = contentType
		}
		if hasBlockEnd {
			break
		}
	}
	if !hasBlockEnd {
		return nil, fmt.Errorf("MIME types block is not closed")
	}
	if len(mimeTypes) == 0 {
		return nil, fmt.Errorf("MIME types block contains no mappings")
	}
	return mimeTypes, nil
}

func setFileContentType(rw http.ResponseWriter, filePath string, mimeTypes map[string]string) {
	contentType := fileContentType(filePath, mimeTypes)
	if contentType == "" {
		// 扩展名未映射：嗅探开头，是文本就直接展示，避免维护扩展名清单。
		if isTextFile(filePath) {
			writePlainText(rw)
			return
		} else {
			contentType = "application/octet-stream"
		}
	} else if !isRenderedType(contentType) && isTextFile(filePath) {
		// 已映射但浏览器不会内联渲染（如 .yaml/.csv/.log）：按纯文本展示。
		writePlainText(rw)
		return
	}
	if strings.HasPrefix(contentType, "text/") {
		contentType += "; charset=utf-8"
	}
	rw.Header().Set("Content-Type", contentType)
}

// writePlainText 以 UTF-8 纯文本返回，浏览器直接内联展示。
func writePlainText(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

// isRenderedType 判断内容类型是否由浏览器内联渲染（而非下载）。
func isRenderedType(contentType string) bool {
	for _, prefix := range renderedTypes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

// isTextFile 嗅探文件开头判断是否为纯文本，结果按扩展名缓存。
// 传入的是 URL 路径，读取时补上文档根目录前缀 "."；读取失败不入缓存，下次重试。
func isTextFile(urlPath string) bool {
	ext := strings.ToLower(filepath.Ext(urlPath))

	textExtCacheMu.RLock()
	cached, ok := textExtCache[ext]
	textExtCacheMu.RUnlock()
	if ok {
		return cached
	}

	file, err := os.Open("." + urlPath)
	if err != nil {
		log.Printf("Cannot sniff %s: %v", urlPath, err)
		return false
	}
	defer file.Close()

	head := make([]byte, textSniffLen)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		log.Printf("Cannot read %s: %v", urlPath, err)
		return false
	}
	isText := strings.HasPrefix(http.DetectContentType(head[:n]), "text/")

	textExtCacheMu.Lock()
	textExtCache[ext] = isText
	textExtCacheMu.Unlock()
	return isText
}

// fileContentType 按扩展名从 mime.types 表取内容类型，未映射时返回空串。
func fileContentType(filePath string, mimeTypes map[string]string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	contentType, ok := mimeTypes[ext]
	if !ok {
		return ""
	}
	return contentType
}

func setDownloadDisposition(rw http.ResponseWriter, filePath string) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".apk", ".apks":
		rw.Header().Set("Content-Disposition", "attachment")
	}
}
