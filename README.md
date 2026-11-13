# markdir

Markdir serves up a directory of files in HTML, rendering Markdown files into HTML when they are encountered as `.md` files.
It's sort of the most degenerate Wiki you could imagine writing short of simply having static HTML files.

## Installation

    go install github.com/weaming/markdir@latest

## Usage

```
Usage of markdir:
  -all
    	show hide directories
  -listen string
    	listen host:port (default "127.0.0.1:10200")
  -no-index string
    	comma separated list of directories to disable listing
  -index-dirs string
        comma separated URL directories to show file details, e.g. '/downloads/,/public/'
  -upload-dirs string
        comma separated URL directories that accept PUT and POST uploads, e.g. '/uploads/'
```

The working directory is the document root. Directories serve `index.html` when present;
otherwise they show a file listing. `-no-index` disables listing but still allows the index page.
The `/date/` virtual routes retain their existing behavior.

Static file MIME types follow the embedded Nginx `mime.types` table; unmapped extensions use `application/octet-stream`.
Files whose first 512 bytes sniff as text are served as `text/plain; charset=utf-8` so browsers display them inline
instead of downloading — source code, configs, `.yaml`, `.csv`, `.log` and extensionless files are all covered.
Types the browser already renders inline (`text/*`, `image/*`, `audio/*`, `video/*`, `font/*`, JSON, XML, JS)
keep their table type so pages, stylesheets and images keep working.
APK and APKS files are served with `Content-Disposition: attachment` to trigger downloads.

`-index-dirs` matches the specified directories and all their subdirectories (`/` selects the entire document tree).
File sizes use 1024-based units with one decimal place, such as `1.0 KiB` or `2.5 MiB`.
Listings show filename, modification date (`YYYY-MM-DD HH:mm`), and right-aligned size;
only filenames are clickable. On mobile, date and size appear right-aligned on the second line.
Dates use `TZ`, defaulting to UTC+8.
Directories and symbolic links show a dash for size. Detailed listings use one column regardless of `-columns`.

`-upload-dirs` enables raw-body PUT and POST uploads to the specified URL directories and their subdirectories.
Missing subdirectories are created; uploads replace existing regular files.

Fenced code blocks tagged `mermaid` are rendered as SVG diagrams (flowchart, sequence, and the rest of Mermaid) in the browser.
Mermaid and the PNG exporter are embedded in the binary and served from `/-/mermaid.min.js` and `/-/html-to-image.js`
with a one-year immutable cache; the mermaid script is only requested by pages that contain diagrams, so diagrams work offline.
Diagrams are re-rendered when the theme changes, and invalid diagram syntax leaves the original code block in place.
Diagram colors follow the page theme (`--diagram-*` in `theme.go`): rounded nodes, hairline edges, and no background box.

### Customization

You can place an `index.css` file in the root directory being served. If present, it will be included in all pages to allow for custom styling.
