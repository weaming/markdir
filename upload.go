package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	urlpath "path"
)

func (r *renderer) serveUpload(rw http.ResponseWriter, req *http.Request) {
	if len(r.uploadDirs) == 0 {
		http.Error(rw, "uploads are disabled", http.StatusForbidden)
		return
	}

	urlPath := req.URL.Path
	cleanURLPath := urlpath.Clean(urlPath)
	if urlPath != cleanURLPath || cleanURLPath == "/" || !isUploadTarget(cleanURLPath, r.uploadDirs) {
		http.Error(rw, "upload path is not allowed", http.StatusForbidden)
		return
	}

	rootPath, err := filepath.Abs(string(r.dir))
	if err != nil {
		http.Error(rw, "cannot resolve document root", http.StatusInternalServerError)
		return
	}
	rootPath, err = filepath.EvalSymlinks(rootPath)
	if err != nil {
		http.Error(rw, "cannot resolve document root", http.StatusInternalServerError)
		return
	}

	relPath := filepath.FromSlash(strings.TrimPrefix(cleanURLPath, "/"))
	if !filepath.IsLocal(relPath) {
		http.Error(rw, "upload path is not allowed", http.StatusForbidden)
		return
	}
	targetPath := filepath.Join(rootPath, relPath)
	parentPath := filepath.Dir(targetPath)
	if err := createUploadDirs(rootPath, parentPath); err != nil {
		http.Error(rw, "cannot create upload directory", http.StatusInternalServerError)
		return
	}

	info, err := os.Lstat(targetPath)
	isNew := errors.Is(err, os.ErrNotExist)
	if err != nil && !isNew {
		http.Error(rw, "cannot access upload target", http.StatusInternalServerError)
		return
	}
	if err == nil && (info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular()) {
		http.Error(rw, "upload target must be a regular file", http.StatusConflict)
		return
	}

	tempFile, err := os.CreateTemp(parentPath, ".markdir-upload-*")
	if err != nil {
		http.Error(rw, "cannot create upload file", http.StatusInternalServerError)
		return
	}
	tempPath := tempFile.Name()
	fileClosed := false
	fileSaved := false
	defer func() {
		if !fileClosed {
			if err := tempFile.Close(); err != nil {
				log.Printf("Cannot close temporary upload %s: %v", tempPath, err)
			}
		}
		if !fileSaved {
			if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Printf("Cannot remove temporary upload %s: %v", tempPath, err)
			}
		}
	}()

	if _, err := io.Copy(tempFile, req.Body); err != nil {
		http.Error(rw, "cannot read uploaded file", http.StatusBadRequest)
		return
	}
	if err := tempFile.Chmod(0644); err != nil {
		http.Error(rw, "cannot set uploaded file permissions", http.StatusInternalServerError)
		return
	}
	if err := tempFile.Close(); err != nil {
		http.Error(rw, "cannot write uploaded file", http.StatusInternalServerError)
		return
	}
	fileClosed = true
	if err := os.Rename(tempPath, targetPath); err != nil {
		http.Error(rw, "cannot save uploaded file", http.StatusInternalServerError)
		return
	}
	fileSaved = true

	if isNew {
		rw.WriteHeader(http.StatusCreated)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}

func isUploadTarget(urlPath string, dirs map[string]bool) bool {
	for dir := range dirs {
		if dir == "/" {
			return urlPath != "/"
		}
		if strings.HasPrefix(urlPath, dir+"/") {
			return true
		}
	}
	return false
}

func createUploadDirs(rootPath, targetPath string) error {
	relPath, err := filepath.Rel(rootPath, targetPath)
	if err != nil {
		return err
	}
	if !filepath.IsLocal(relPath) {
		return errors.New("upload directory is outside the document root")
	}
	if relPath == "." {
		return nil
	}

	currentPath := rootPath
	for _, part := range strings.Split(relPath, string(filepath.Separator)) {
		currentPath = filepath.Join(currentPath, part)
		info, err := os.Lstat(currentPath)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(currentPath, 0755); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			info, err = os.Lstat(currentPath)
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("upload directory contains a symbolic link")
		}
		if !info.IsDir() {
			return errors.New("upload path component is not a directory")
		}
	}
	return nil
}
