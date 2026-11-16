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

const directoryUploadHTML = `
<style>
.upload-tools{display:flex;align-items:center;gap:12px;margin-bottom:12px}
.upload-button{flex:0 0 auto;white-space:nowrap;font:inherit;font-size:14px;min-height:44px;padding:4px 12px;border:1px solid var(--btn-border);border-radius:6px;color:var(--link);background:var(--card-bg);cursor:pointer}
.upload-button:disabled{opacity:.6;cursor:wait}
.upload-button:focus-visible{outline:2px solid var(--link);outline-offset:2px}
.upload-status{flex:1;min-width:0;font-size:14px;overflow-wrap:anywhere}
.upload-status:empty{display:none}
@media(max-width:600px){.upload-tools{flex-wrap:wrap;gap:8px;padding:0 16px}.upload-status{flex-basis:100%}}
@media print{.upload-tools{display:none}}
</style>
<div class="upload-tools">
  <button type="button" class="upload-button" id="upload-button">↑ 上传</button>
  <input type="file" id="upload-files" multiple hidden>
  <span class="upload-status" id="upload-status" role="status" aria-live="polite"></span>
</div>
<script>
(function () {
  const button = document.getElementById('upload-button');
  const picker = document.getElementById('upload-files');
  const status = document.getElementById('upload-status');
  button.addEventListener('click', function () { picker.click(); });
  picker.addEventListener('change', async function () {
    const files = Array.from(picker.files);
    if (!files.length) return;
    button.disabled = true;
    let completed = 0;
    try {
      for (const file of files) {
        if (!file.name || file.name === '.' || file.name === '..' || /[/\\]/.test(file.name)) {
          throw new Error('无效文件名');
        }
        status.textContent = '上传 ' + (completed + 1) + '/' + files.length + '：' + file.name;
        const target = new URL(window.location.href);
        target.search = '';
        target.hash = '';
        target.pathname = target.pathname.replace(/\/?$/, '/') + encodeURIComponent(file.name);
        await new Promise(function (resolve, reject) {
          const request = new XMLHttpRequest();
          request.open('PUT', target.href);
          request.upload.onprogress = function (event) {
            if (event.lengthComputable) {
              status.textContent = '上传 ' + (completed + 1) + '/' + files.length + '：' + file.name + ' ' + Math.round(event.loaded / event.total * 100) + '%';
            }
          };
          request.onload = function () {
            if (request.status >= 200 && request.status < 300) resolve();
            else reject(new Error(file.name + '：' + (request.responseText.trim() || 'HTTP ' + request.status)));
          };
          request.onerror = function () { reject(new Error(file.name + '：网络错误，请重试')); };
          request.onabort = function () { reject(new Error(file.name + '：上传已取消')); };
          request.send(file);
        });
        completed++;
      }
      window.location.reload();
    } catch (error) {
      status.textContent = '已上传 ' + completed + '/' + files.length + '；' + error.message;
    } finally {
      button.disabled = false;
      picker.value = '';
    }
  });
})();
</script>
`

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
