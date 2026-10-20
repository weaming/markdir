package main

const MDTemplate = `
<html>
  <head>
	<meta http-equiv="content-Type" content="text/html; charset=UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>{{ .Path }}</title>
    ` + themeHeadScript + `
    ` + imageExportHead + `
    {{ if .HasCustomCSS }}<link rel="stylesheet" href="/index.css">{{ end }}
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
      ` + imageExportCSS + `

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

      .markdown-body > *:first-child {
        margin-top: 0 !important;
      }

      .markdown-body > *:last-child {
        margin-bottom: 0 !important;
      }

      .markdown-body a {
        color: var(--link);
        text-decoration: underline;
        text-decoration-color: color-mix(in srgb, currentColor 35%, transparent);
        text-decoration-thickness: 1px;
        text-underline-offset: 0.18em;
        transition: color 0.2s ease, text-decoration-color 0.2s ease;
      }

      .markdown-body a:hover {
        color: var(--link-hover);
        text-decoration-color: currentColor;
      }

      .markdown-body a:visited {
        color: var(--link-visited);
      }

      .markdown-body a:visited:hover {
        color: var(--link-visited-hover);
        text-decoration-color: currentColor;
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
        border-bottom: 1px solid var(--heading-border);
      }

      .markdown-body h2 {
        padding-bottom: 0.25em;
        font-size: 1.45em;
        border-bottom: 1px solid var(--heading-border);
      }

      .markdown-body h3 {
        margin-top: 28px;
        font-size: 1.2em;
      }

      .markdown-body h4 {
        margin-top: 24px;
        font-size: 1.05em;
      }

      .markdown-body h5 {
        margin-top: 22px;
        font-size: 1em;
      }

      .markdown-body h6 {
        margin-top: 20px;
        font-size: 0.9em;
        color: var(--muted);
      }

      .markdown-body p {
        margin: 0 0 16px;
      }

      .markdown-body ul,
      .markdown-body ol {
        padding-left: 1.5em;
        margin: 0 0 16px;
      }

      .markdown-body li {
        margin-bottom: 4px;
      }

      .markdown-body li > ul,
      .markdown-body li > ol {
        margin-top: 4px;
        margin-bottom: 4px;
      }

      .markdown-body li:has(> input[type="checkbox"]) {
        list-style: none;
      }

      .markdown-body li > input[type="checkbox"] {
        margin: 0 0.4em 0 0;
        vertical-align: middle;
      }

      .markdown-body blockquote {
        padding: 0 12px;
        margin: 0 0 16px;
        color: var(--muted);
        border-left: 4px solid var(--quote-border);
      }

      .markdown-body blockquote > :first-child {
        margin-top: 0;
      }

      .markdown-body blockquote > :last-child {
        margin-bottom: 0;
      }

      .markdown-body code,
      .markdown-body pre {
        font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Monaco,
          Consolas, "DejaVu Sans Mono", monospace;
        letter-spacing: normal;
        font-variant-ligatures: common-ligatures discretionary-ligatures contextual;
        text-autospace: no-autospace;
      }

      .markdown-body code {
        padding: 0.2em 0.4em;
        margin: 0;
        font-size: 85%;
        background-color: var(--code-bg);
        border-radius: 3px;
      }

      .markdown-body pre {
        padding: 12px;
        margin: 0 0 16px;
        overflow: auto;
        font-size: 85%;
        line-height: 1.275;
        background-color: var(--pre-bg);
        border-radius: 6px;
      }

      .markdown-body pre code {
        padding: 0;
        margin: 0;
        background: transparent;
        border: 0;
        font-size: 100%;
      }

      .markdown-body pre code .code-cell {
        display: inline-block;
        width: 1ch;
        text-align: center;
        vertical-align: baseline;
      }

      .markdown-body pre code .code-cell[data-cols="2"] {
        width: 2ch;
      }

      .markdown-body table {
        display: block;
        width: 100%;
        overflow: auto;
        margin: 0 0 16px;
        border-collapse: collapse;
        font-variant-numeric: tabular-nums;
      }

      .markdown-body table th,
      .markdown-body table td {
        padding: 5px 10px;
        border: 1px solid var(--table-border);
      }

      .markdown-body table th {
        font-weight: 600;
        background-color: var(--table-alt-bg);
      }

      .markdown-body table tr {
        background-color: var(--card-bg);
        border-top: 1px solid var(--table-row-border);
      }

      .markdown-body table tr:nth-child(2n) {
        background-color: var(--table-alt-bg);
      }

      .markdown-body img {
        max-width: 100%;
        box-sizing: content-box;
      }

      .markdown-body hr {
        height: 0.25em;
        padding: 0;
        margin: 16px 0;
        background-color: var(--hr-bg);
        border: 0;
      }

      .markdown-body kbd {
        display: inline-block;
        padding: 3px 5px;
        font-size: 11px;
        line-height: 10px;
        color: var(--kbd-fg);
        vertical-align: middle;
        background-color: var(--bg);
        border: 1px solid var(--kbd-border);
        border-radius: 3px;
        box-shadow: inset 0 -1px 0 var(--kbd-border);
        font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace;
      }

      .markdown-body mark {
        background: var(--mark-bg);
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

      ` + themeToggleCSS + `

      @media (max-width: 600px) {
        body {
          padding: 0;
          font-size: 16px;
          line-height: 1.65;
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
          margin-top: 18px;
          margin-bottom: 5px;
        }

        .markdown-body h4,
        .markdown-body h5,
        .markdown-body h6 {
          font-size: 1em;
          margin-top: 16px;
          margin-bottom: 4px;
        }

        .markdown-body p {
          margin: 0 0 14px;
        }

        .markdown-body ul,
        .markdown-body ol {
          padding-left: 1.6em;
          margin: 0 0 14px;
        }

        .markdown-body li {
          margin-bottom: 3px;
        }

        .markdown-body pre {
          padding: 12px;
          margin: 0 -16px 14px;
          border-radius: 0;
          font-size: 13px;
          line-height: 1.5;
        }

        .markdown-body code {
          font-size: 13px;
        }

        .markdown-body blockquote {
          margin: 0 0 14px;
          padding: 0 12px;
        }

        .markdown-body table {
          display: block;
          overflow-x: auto;
          margin: 0 0 14px;
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

        .theme-toggle {
          display: none;
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
    ` + themeButton + `
    ` + imageExportButton + `
    <div class="markdown-body">
      {{.Body}}
    </div>
  ` + themeToggleScript + `
  ` + imageExportScript + `
  ` + codeCellScript + `
  </body>
</html>`

const codeCellScript = `
<script>
  (function () {
    if (!Intl.Segmenter) return;

    const NON_ASCII = /[^\x00-\x7f]/;
    const EMOJI_PRESENTATION = /\p{Emoji_Presentation}/u;
    const SEGMENTER = new Intl.Segmenter(undefined, { granularity: "grapheme" });
    // Iosevka Code 中这些字符宽 1000 单位，ASCII 字符宽 500 单位。
    const CODE_DOUBLE_WIDTH_RANGES = [
      [0x2001, 0x2001], [0x2003, 0x2003], [0x2014, 0x2015], [0x2024, 0x2026],
      [0x2030, 0x2031], [0x203b, 0x203b], [0x2042, 0x2042], [0x2053, 0x2053],
      [0x212e, 0x212e], [0x2140, 0x2140], [0x214f, 0x214f],
      [0x220f, 0x2211], [0x221d, 0x221e], [0x223e, 0x223e], [0x22c0, 0x22c3],
      [0x22ee, 0x22f1], [0x2301, 0x2301], [0x2303, 0x2304], [0x230c, 0x230f],
      [0x2311, 0x2318], [0x231a, 0x231f], [0x2324, 0x2328], [0x232b, 0x232b],
      [0x232d, 0x2335], [0x237b, 0x237b], [0x2384, 0x238b], [0x2394, 0x2394],
      [0x23b2, 0x23b3], [0x23c0, 0x23ca], [0x23ce, 0x23cf], [0x23e2, 0x23e2],
      [0x23e4, 0x23e5], [0x23e9, 0x23fe], [0x25a0, 0x25cb],
      [0x25cd, 0x25e5], [0x25e7, 0x2600], [0x2605, 0x2607], [0x2609, 0x2609],
      [0x2610, 0x2613], [0x261c, 0x261f], [0x2626, 0x2626], [0x2628, 0x2628],
      [0x2639, 0x263c], [0x267e, 0x267e], [0x2680, 0x2689], [0x2699, 0x2699],
      [0x269b, 0x269b], [0x26aa, 0x26af], [0x26cb, 0x26cb], [0x26da, 0x26da],
      [0x26dd, 0x26de], [0x26ed, 0x26ef], [0x26f6, 0x26f6], [0x2705, 0x2705],
      [0x2713, 0x2718], [0x271a, 0x271d], [0x2720, 0x2727], [0x2731, 0x2734],
      [0x2736, 0x274c], [0x274e, 0x274e], [0x2756, 0x2756], [0x2758, 0x275a],
      [0x2764, 0x2764], [0x2776, 0x2794], [0x279c, 0x279e], [0x27a1, 0x27a7],
      [0x27b0, 0x27b0], [0x27bf, 0x27bf], [0x27cb, 0x27cb], [0x27cd, 0x27cd],
      [0x27d8, 0x27d9], [0x27dd, 0x27de], [0x27e0, 0x27e1], [0x27f0, 0x27ff],
      [0x2900, 0x2926], [0x2933, 0x2933], [0x2940, 0x2941], [0x2947, 0x2948],
      [0x294a, 0x2965], [0x296e, 0x296f], [0x2981, 0x2981], [0x29c9, 0x29cb],
      [0x29cd, 0x29cd], [0x29e8, 0x29e9], [0x29eb, 0x29eb], [0x2a00, 0x2a02],
      [0x2a05, 0x2a06], [0x2a09, 0x2a09], [0x2a1d, 0x2a1e], [0x2b00, 0x2b19],
      [0x2b1b, 0x2b3f], [0x2b45, 0x2b46], [0x2b4d, 0x2b4d], [0x2b50, 0x2b55],
      [0x2b58, 0x2b59], [0x2b60, 0x2b73], [0x2b7a, 0x2b87], [0x2b8c, 0x2b95],
      [0x2ba0, 0x2ba7], [0x2bb8, 0x2bc8], [0x2bca, 0x2bd0], [0x2be8, 0x2bef],
      [0x2e43, 0x2e43], [0x3248, 0x324f], [0x3251, 0x325f], [0x32b1, 0x32bf],
      [0xee06, 0xee0b], [0xf881, 0xf881], [0xfffc, 0xfffc], [0x1cc69, 0x1cc6f],
      [0x1cc89, 0x1cc8f], [0x1ce00, 0x1ce01], [0x1ce03, 0x1ce06], [0x1ce0b, 0x1ce0c],
      [0x1f100, 0x1f10c], [0x1f110, 0x1f12e], [0x1f130, 0x1f169], [0x1f170, 0x1f1ac],
      [0x1f311, 0x1f318], [0x1f377, 0x1f377], [0x1f514, 0x1f514], [0x1f547, 0x1f548],
      [0x1f550, 0x1f567], [0x1f5a4, 0x1f5a4], [0x1f610, 0x1f610], [0x1f636, 0x1f636],
      [0x1f641, 0x1f643], [0x1f780, 0x1f7bf], [0x1f800, 0x1f80b], [0x1f810, 0x1f847],
      [0x1f850, 0x1f859], [0x1f860, 0x1f887], [0x1f8b0, 0x1f8bb], [0x1f8c0, 0x1f8c1],
      [0x1f8d0, 0x1f8d8], [0x1f907, 0x1f907], [0x1f90d, 0x1f90d], [0x1fbe0, 0x1fbe3],
      [0x1fbe8, 0x1fbef],
    ];

    function isCodeDoubleWidth(codePoint) {
      let left = 0;
      let right = CODE_DOUBLE_WIDTH_RANGES.length - 1;

      while (left <= right) {
        const middle = Math.floor((left + right) / 2);
        const [start, end] = CODE_DOUBLE_WIDTH_RANGES[middle];

        if (codePoint < start) {
          right = middle - 1;
        } else if (codePoint > end) {
          left = middle + 1;
        } else {
          return true;
        }
      }

      return false;
    }

    function isWideCell(segment) {
      const codePoint = segment.codePointAt(0);
      return isCodeDoubleWidth(codePoint) ||
        (codePoint >= 0x1100 && codePoint <= 0x115f) ||
        codePoint === 0x2329 || codePoint === 0x232a ||
        (codePoint >= 0x2e80 && codePoint <= 0xa4cf) ||
        (codePoint >= 0xac00 && codePoint <= 0xd7a3) ||
        (codePoint >= 0xf900 && codePoint <= 0xfaff) ||
        (codePoint >= 0xfe10 && codePoint <= 0xfe6f) ||
        (codePoint >= 0xff01 && codePoint <= 0xff60) ||
        (codePoint >= 0xffe0 && codePoint <= 0xffe6) ||
        EMOJI_PRESENTATION.test(segment) || segment.includes("\ufe0f");
    }

    function alignTextNode(textNode) {
      const fragment = document.createDocumentFragment();
      let plainText = "";

      for (const { segment } of SEGMENTER.segment(textNode.textContent)) {
        if (!NON_ASCII.test(segment)) {
          plainText += segment;
          continue;
        }

        if (plainText) {
          fragment.append(document.createTextNode(plainText));
          plainText = "";
        }

        const cell = document.createElement("span");
        cell.className = "code-cell";
        if (isWideCell(segment)) cell.dataset.cols = "2";
        cell.textContent = segment;
        fragment.append(cell);
      }

      if (plainText) fragment.append(document.createTextNode(plainText));
      textNode.replaceWith(fragment);
    }

    for (const codeBlock of document.querySelectorAll(".markdown-body pre code")) {
      const walker = document.createTreeWalker(codeBlock, NodeFilter.SHOW_TEXT);
      const textNodes = [];

      while (walker.nextNode()) {
        if (NON_ASCII.test(walker.currentNode.textContent)) {
          textNodes.push(walker.currentNode);
        }
      }

      for (const textNode of textNodes) alignTextNode(textNode);
    }
  })();
</script>`
