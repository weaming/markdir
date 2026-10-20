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
  -size-dirs string
        comma separated URL directories to show file sizes, e.g. '/downloads/,/public/'
```

The working directory is the document root. Directories serve `index.html` when present;
otherwise they show a file listing. `-no-index` disables listing but still allows the index page.
The `/date/` virtual routes retain their existing behavior.

Static file MIME types follow the embedded Nginx `mime.types` table; unknown extensions use `application/octet-stream`.

`-size-dirs` matches only the specified directories, without recursion (`/` selects the root).
File sizes use 1024-based units with one decimal place, such as `1.0 KiB` or `2.5 MiB`.
Listings show filename, modification date (`YYYY-MM-DD HH:mm`), and right-aligned size;
only filenames are clickable. Dates use `TZ`, defaulting to UTC+8.
Directories and symbolic links show a dash for size. Detailed listings use one column regardless of `-columns`.

### Customization

You can place an `index.css` file in the root directory being served. If present, it will be included in all pages to allow for custom styling.
