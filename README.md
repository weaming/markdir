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
```

### Customization

You can place an `index.css` file in the root directory being served. If present, it will be included in all pages to allow for custom styling.
