package main

const MDTemplate = `
<html>
  <head>
	<meta http-equiv="content-Type" content="text/html; charset=UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{ .Path }}</title>
    {{ if .HasCustomCSS }}<link rel="stylesheet" href="/index.css">{{ end }}
    <style>
      * {
        -webkit-box-sizing: border-box;
        box-sizing: border-box;
      }

      body {
        margin: 0;
        padding: 20px;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
        font-size: 16px;
        line-height: 1.5;
        color: #24292e;
        background: #fafbfc;
      }

      .markdown-body {
        max-width: 1000px;
        margin: 0 auto;
        padding: 24px 32px;
        background: #fff;
        border-radius: 6px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.08);
      }

      .markdown-body > *:first-child {
        margin-top: 0 !important;
      }

      .markdown-body > *:last-child {
        margin-bottom: 0 !important;
      }

      .markdown-body a {
        color: #0366d6;
        text-decoration: none;
        border-bottom: 1px solid transparent;
        transition: color 0.2s ease, border-color 0.2s ease;
      }

      .markdown-body a:hover {
        color: #0256b9;
        border-bottom-color: #0256b9;
        text-decoration: none;
      }

      .markdown-body a:visited {
        color: #8b5cf6;
      }

      .markdown-body a:visited:hover {
        color: #6d42e6;
        border-bottom-color: #6d42e6;
      }

      .markdown-body h1,
      .markdown-body h2,
      .markdown-body h3,
      .markdown-body h4,
      .markdown-body h5,
      .markdown-body h6 {
        margin-top: 16px;
        margin-bottom: 8px;
        font-weight: 600;
        line-height: 1.25;
      }

      .markdown-body h1 {
        padding-bottom: 0.25em;
        font-size: 1.75em;
        border-bottom: 1px solid #eaecef;
      }

      .markdown-body h2 {
        padding-bottom: 0.25em;
        font-size: 1.45em;
        border-bottom: 1px solid #eaecef;
      }

      .markdown-body h3 {
        font-size: 1.2em;
      }

      .markdown-body h4 {
        font-size: 1.05em;
      }

      .markdown-body h5 {
        font-size: 1em;
      }

      .markdown-body h6 {
        font-size: 0.9em;
        color: #6a737d;
      }

      .markdown-body p {
        margin-bottom: 8px;
      }

      .markdown-body ul,
      .markdown-body ol {
        padding-left: 1.5em;
        margin-top: 0;
        margin-bottom: 4px;
      }

      .markdown-body li {
        margin-bottom: 4px;
      }

      .markdown-body li > ul,
      .markdown-body li > ol {
        margin-top: 4px;
        margin-bottom: 4px;
      }

      .markdown-body ol + ul,
      .markdown-body ul + ol,
      .markdown-body ol + ol,
      .markdown-body ul + ul {
        margin-top: -4px;
      }

      .markdown-body blockquote {
        padding: 0 12px;
        margin: 0 0 8px;
        color: #6a737d;
        border-left: 4px solid #dfe2e5;
      }

      .markdown-body blockquote > :first-child {
        margin-top: 0;
      }

      .markdown-body blockquote > :last-child {
        margin-bottom: 0;
      }

      .markdown-body code {
        padding: 0.2em 0.4em;
        margin: 0;
        font-size: 85%;
        background-color: rgba(27,31,35,0.05);
        border-radius: 3px;
        font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace;
      }

      .markdown-body pre {
        padding: 12px;
        margin-bottom: 8px;
        overflow: auto;
        font-size: 85%;
        line-height: 1.4;
        background-color: #f6f8fa;
        border-radius: 6px;
        font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace;
      }

      .markdown-body pre code {
        padding: 0;
        margin: 0;
        background: transparent;
        border: 0;
        font-size: 100%;
      }

      .markdown-body table {
        display: block;
        width: 100%;
        overflow: auto;
        margin-bottom: 8px;
        border-collapse: collapse;
      }

      .markdown-body table th,
      .markdown-body table td {
        padding: 5px 10px;
        border: 1px solid #dfe2e5;
      }

      .markdown-body table th {
        font-weight: 600;
        background-color: #f6f8fa;
      }

      .markdown-body table tr {
        background-color: #fff;
        border-top: 1px solid #c6cbd1;
      }

      .markdown-body table tr:nth-child(2n) {
        background-color: #f6f8fa;
      }

      .markdown-body img {
        max-width: 100%;
        box-sizing: content-box;
      }

      .markdown-body hr {
        height: 0.25em;
        padding: 0;
        margin: 16px 0;
        background-color: #e1e4e8;
        border: 0;
      }

      .markdown-body kbd {
        display: inline-block;
        padding: 3px 5px;
        font-size: 11px;
        line-height: 10px;
        color: #444d56;
        vertical-align: middle;
        background-color: #fafbfc;
        border: 1px solid #d1d5da;
        border-radius: 3px;
        box-shadow: inset 0 -1px 0 #d1d5da;
        font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace;
      }

      .markdown-body mark {
        background: #fff3c4;
        color: inherit;
        padding: 0.2em 0.4em;
        border-radius: 3px;
      }

      .markdown-body strong {
        font-weight: 600;
      }

      .markdown-body del {
        text-decoration: line-through;
      }

      @media (max-width: 600px) {
        body {
          padding: 0;
          font-size: 15px;
          line-height: 1.55;
        }

        .markdown-body {
          padding: 18px 16px;
          border-radius: 0;
          box-shadow: none;
        }

        .markdown-body h1 {
          font-size: 1.5em;
          margin-top: 12px;
          margin-bottom: 6px;
        }

        .markdown-body h2 {
          font-size: 1.3em;
          margin-top: 12px;
          margin-bottom: 5px;
        }

        .markdown-body h3 {
          font-size: 1.15em;
          margin-top: 10px;
          margin-bottom: 5px;
        }

        .markdown-body h4,
        .markdown-body h5,
        .markdown-body h6 {
          font-size: 1em;
          margin-top: 8px;
          margin-bottom: 4px;
        }

        .markdown-body p {
          margin-bottom: 10px;
        }

        .markdown-body ul,
        .markdown-body ol {
          padding-left: 1.6em;
          margin-bottom: 8px;
        }

        .markdown-body li {
          margin-bottom: 3px;
        }

        .markdown-body pre {
          padding: 12px;
          margin: 0 -16px 12px;
          border-radius: 0;
          font-size: 13px;
          line-height: 1.45;
        }

        .markdown-body code {
          font-size: 13px;
        }

        .markdown-body blockquote {
          margin: 0 0 10px;
          padding: 0 12px;
        }

        .markdown-body table {
          display: block;
          overflow-x: auto;
          margin-bottom: 10px;
          font-size: 14px;
        }

        .markdown-body hr {
          margin: 14px 0;
        }

        .markdown-body img {
          max-width: 100%;
          height: auto;
        }
      }

      @media print {
        body {
          background: #fff;
          padding: 0;
          font-size: 10pt;
          line-height: 1.4;
          color: #000;
        }

        .markdown-body {
          box-shadow: none;
          max-width: 100%;
          padding: 15pt 0;
          border-radius: 0;
        }

        .markdown-body a {
          color: inherit;
          text-decoration: underline;
        }

        .markdown-body a[href]:after {
          content: " (" attr(href) ")";
          font-size: 8pt;
          color: #666;
        }

        .markdown-body a[href^="#"]:after,
        .markdown-body a[href^="javascript:"]:after {
          content: "";
        }

        .markdown-body pre,
        .markdown-body blockquote {
          page-break-inside: avoid;
          border: 1px solid #ddd;
          background: #f9f9f9;
        }

        .markdown-body pre {
          white-space: pre-wrap;
          word-wrap: break-word;
        }

        .markdown-body h1,
        .markdown-body h2,
        .markdown-body h3 {
          page-break-after: avoid;
        }

        .markdown-body img {
          max-width: 100% !important;
          page-break-inside: avoid;
        }

        .markdown-body table {
          page-break-inside: avoid;
        }

        .markdown-body tr {
          page-break-inside: avoid;
        }

        .markdown-body h1,
        .markdown-body h2,
        .markdown-body h3,
        .markdown-body h4,
        .markdown-body h5,
        .markdown-body h6 {
          margin-top: 10pt;
          margin-bottom: 4pt;
        }

        .markdown-body p {
          margin-bottom: 6pt;
          orphans: 2;
          widows: 2;
        }

        .markdown-body ul,
        .markdown-body ol {
          margin-bottom: 6pt;
        }
      }
    </style>
  </head>
  <body>
    <div class="markdown-body">
      {{.Body}}
    </div>
  </body>
</html>`
