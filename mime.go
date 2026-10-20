package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed mime.types
var nginxMIMETypes string

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
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filePath)), ".")
	contentType, ok := mimeTypes[ext]
	if !ok {
		contentType = "application/octet-stream"
	}
	rw.Header().Set("Content-Type", contentType)
}

func setDownloadDisposition(rw http.ResponseWriter, filePath string) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".apk", ".apks":
		rw.Header().Set("Content-Disposition", "attachment")
	}
}
