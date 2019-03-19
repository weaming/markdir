package main

const MDTemplateIndex = `
<html>
  <head>
	<meta http-equiv="content-Type" content="text/html; charset=UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Directory</title>
    <style>
      body {
        max-width: 980px;
        border: 1px solid #ddd;
        outline: 1300px solid #fff;
        margin: 16px auto;
      }

      .markdown-body {
        padding: 45px;
        font-family: sans-serif;
        -ms-text-size-adjust: 100%;
        -webkit-text-size-adjust: 100%;
        color: #333333;
        overflow: hidden;
        font-family: "Helvetica Neue", Helvetica, "Segoe UI", Arial, freesans, sans-serif;
        font-size: 16px;
        line-height: 1.6;
        word-wrap: break-word;
      }

	  .markdown-body * {
		-webkit-box-sizing: border-box;
		box-sizing: border-box;
	  }

      .markdown-body a {
        background: transparent;
        color: #4183c4;
        text-decoration: none;
      }

	  .markdown-body a:visited {
		color: #85C1E9;
	  }

      .markdown-body a:active,
      .markdown-body a:hover {
        outline: 0;
      }

      .markdown-body a:hover,
      .markdown-body a:focus,
      .markdown-body a:active {
        text-decoration: underline;
      }

      @media print {
        .markdown-body * {
          background: transparent !important;
          color: black !important;
          filter: none !important;
          -ms-filter: none !important;
        }

        .markdown-body {
          font-size: 12pt;
          max-width: 100%;
          outline: none;
          border: 0;
        }

        .markdown-body a,
        .markdown-body a:visited {
          text-decoration: underline;
        }

        .markdown-body a[href]:after {
          content: " (" attr(href) ")";
        }

        .markdown-body a[href^="javascript:"]:after,
        .markdown-body a[href^="#"]:after {
          content: "";
        }

        .markdown-body pre {
          white-space: pre;
          white-space: pre-wrap;
          word-wrap: break-word;
          border: 1px solid #999;
          padding-right: 1em;
          page-break-inside: avoid;
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
