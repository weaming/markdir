package main

const MDTemplateIndex = `
<html>
  <head>
	<meta http-equiv="content-Type" content="text/html; charset=UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{TITLE}}</title>
    ` + themeHeadScript + `
    <style>
      * {
        -webkit-box-sizing: border-box;
        box-sizing: border-box;
      }

      :root {` + themeLightVars + `}

      :root[data-theme="dark"] {` + themeDarkVars + `}

      @media (prefers-color-scheme: dark) {
      :root:not([data-theme="light"]) {` + themeDarkVars + `}
      }

      ` + themeDarkTypography + `
      ` + themeSelection + `

      body {
        margin: 0;
        padding: 20px;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
        font-size: 17px;
        line-height: 1.7;
        text-autospace: normal;
        color: var(--fg);
        background: var(--bg);
      }

      .markdown-body {
        max-width: 900px;
        margin: 0 auto;
        padding: 24px 32px;
        background: var(--card-bg);
        border-radius: 6px;
        box-shadow: var(--card-shadow);
        overflow-wrap: break-word;
      }

      .dir-list {
        list-style: none;
        margin: 0;
        padding: 0;
      }

      .dir-list li {
        margin-bottom: 10px;
      }

      .dir-list a {
        font-size: 20px;
        color: var(--link);
        text-decoration: none;
        border-bottom: 1px solid transparent;
        transition: color 0.2s ease, border-color 0.2s ease;
      }

      .dir-list a:hover {
        color: var(--link-hover);
        border-bottom-color: var(--link-hover);
      }

      .dir-details li {
        display: flex;
        align-items: baseline;
        gap: 16px;
      }

      .dir-details a {
        flex: 0 1 auto;
        min-width: 0;
        overflow-wrap: anywhere;
        border-bottom: 0;
      }

      .dir-details a:hover {
        text-decoration: underline;
      }

      .file-details {
        display: grid;
        grid-template-columns: 16ch 10ch;
        gap: 16px;
        margin-left: auto;
      }

      .file-modified, .file-size {
        font-family: ui-monospace, monospace;
        font-size: 14px;
        font-variant-numeric: tabular-nums;
        color: var(--fg);
        white-space: nowrap;
      }

      .file-size {
        text-align: right;
      }

      .dir-list a:visited {
        color: var(--link-visited);
      }

      .dir-list a:visited:hover {
        color: var(--link-visited-hover);
        border-bottom-color: var(--link-visited-hover);
      }

      ` + themeToggleCSS + `

      @media (max-width: 600px) {
        body {
          padding: 0;
          font-size: 16px;
          line-height: 1.65;
        }

        .markdown-body {
          padding: 12px 0;
          border-radius: 0;
          box-shadow: none;
        }

        .dir-list li {
          border-bottom: 1px solid var(--divider);
        }

        .dir-list li:last-child {
          border-bottom: none;
        }

        .dir-list li {
          margin-bottom: 0;
        }

        .dir-list a {
          display: block;
          padding: 14px 16px;
          font-size: 17px;
          border-bottom: none;
          transition: background 0.15s ease, color 0.15s ease;
        }

        .dir-list a:hover {
          background: var(--hover-light-bg);
          color: var(--link-hover);
          border-bottom: none;
        }

        .dir-list a:visited:hover {
          background: var(--hover-purple-bg);
          color: var(--link-visited-hover);
          border-bottom: none;
        }

        .dir-details li {
          gap: 8px;
        }

        .dir-details a {
          flex: 1 1 auto;
          padding: 14px 0 14px 16px;
        }

        .file-details {
          grid-template-columns: auto auto;
          gap: 8px;
          padding-right: 16px;
        }
      }

      @media print {
        body {
          background: #fff;
          padding: 0;
        }

        .theme-toggle {
          display: none;
        }

        .markdown-body {
          box-shadow: none;
          max-width: 100%;
          padding: 20px 0;
          border-radius: 0;
        }
      }
    </style>
  </head>
  <body>
    ` + themeButton + `
    <div class="markdown-body">
	`

const MDTemplateIndexTail = `
	</div>
  ` + themeToggleScript + `
  </body>
</html>`
