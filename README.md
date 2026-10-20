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

Static file MIME types follow the embedded Nginx `mime.types` table; unknown extensions use `application/octet-stream`.
APK and APKS files are served with `Content-Disposition: attachment` to trigger downloads.

`-index-dirs` matches the specified directories and all their subdirectories (`/` selects the entire document tree).
File sizes use 1024-based units with one decimal place, such as `1.0 KiB` or `2.5 MiB`.
Listings show filename, modification date (`YYYY-MM-DD HH:mm`), and right-aligned size;
only filenames are clickable. On mobile, date and size appear right-aligned on the second line.
Dates use `TZ`, defaulting to UTC+8.
Directories and symbolic links show a dash for size. Detailed listings use one column regardless of `-columns`.

`-upload-dirs` enables raw-body PUT and POST uploads to the specified URL directories and their subdirectories.
Missing subdirectories are created; uploads replace existing regular files.

### Customization

You can place an `index.css` file in the root directory being served. If present, it will be included in all pages to allow for custom styling.
