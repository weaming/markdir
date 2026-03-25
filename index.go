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
        max-width: 1100px;
        margin: 0 auto;
        padding: 30px 45px;
        background: #fff;
        border-radius: 8px;
        box-shadow: 0 1px 3px rgba(0,0,0,0.08);
      }

      .markdown-body a {
        color: #0366d6;
        text-decoration: none;
      }

      .markdown-body a:hover {
        text-decoration: underline;
      }

      .markdown-body a:visited {
        color: #0366d6;
      }

      @media (max-width: 600px) {
        body {
          padding: 0;
          font-size: 15px;
        }

        .markdown-body {
          padding: 18px 16px;
          border-radius: 0;
          box-shadow: none;
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
