package main

const MDTemplateIndex = `
<html>
  <head>
	<meta http-equiv="content-Type" content="text/html; charset=UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Directory</title>
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
        color: #0366d6;
        text-decoration: none;
        border-bottom: 1px solid transparent;
        transition: color 0.2s ease, border-color 0.2s ease;
      }

      .dir-list a:hover {
        color: #0256b9;
        border-bottom-color: #0256b9;
      }

      .dir-list a:visited {
        color: #8b5cf6;
      }

      .dir-list a:visited:hover {
        color: #6d42e6;
        border-bottom-color: #6d42e6;
      }

      @media (max-width: 600px) {
        body {
          padding: 0;
          font-size: 15px;
        }

        .markdown-body {
          padding: 12px 0;
          border-radius: 0;
          box-shadow: none;
        }

        .dir-list li {
          border-bottom: 1px solid #f0f0f0;
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
          background: #f0f6ff;
          color: #0256b9;
          border-bottom: none;
        }

        .dir-list a:visited:hover {
          background: #f5f0ff;
          color: #6d42e6;
          border-bottom: none;
        }
      }

      @media print {
        body {
          background: #fff;
          padding: 0;
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
    <div class="markdown-body">
	`

const MDTemplateIndexTail = `
	</div>
  </body>
</html>`
