package main

// Shared theme (dark mode) support for both HTML templates: palettes as CSS
// variables, a floating toggle button, and inline scripts. The button cycles
// auto → light → dark and persists the choice in localStorage; "auto" (or no
// saved value) follows the system preference via the CSS media query.

const themeLightVars = `
        color-scheme: light;
        --bg: #fafbfc;
        --fg: #24292e;
        --card-bg: #fff;
        --card-shadow: 0 1px 3px rgba(0,0,0,0.08);
        --link: #0366d6;
        --link-hover: #0256b9;
        --link-visited: #8250df;
        --link-visited-hover: #6d42e6;
        --heading-border: #eaecef;
        --muted: #57606a;
        --quote-border: #dfe2e5;
        --code-bg: rgba(27,31,35,0.05);
        --pre-bg: #f6f8fa;
        --table-border: #dfe2e5;
        --table-row-border: #c6cbd1;
        --table-alt-bg: #f6f8fa;
        --hr-bg: #e1e4e8;
        --kbd-fg: #444d56;
        --kbd-border: #d1d5da;
        --mark-bg: #fff3c4;
        --btn-border: #d1d5da;
        --hover-light-bg: #f0f6ff;
        --hover-purple-bg: #f5f0ff;
        --divider: #f0f0f0;
        --selection-bg: rgba(3,102,214,0.18);
      `

const themeDarkVars = `
        color-scheme: dark;
        --bg: #0d1117;
        --fg: #d8dee4;
        --card-bg: #161b22;
        --card-shadow: 0 1px 3px rgba(0,0,0,0.45);
        --link: #58a6ff;
        --link-hover: #79c0ff;
        --link-visited: #bc8cff;
        --link-visited-hover: #d2a8ff;
        --heading-border: #30363d;
        --muted: #8b949e;
        --quote-border: #30363d;
        --code-bg: rgba(110,118,129,0.25);
        --pre-bg: #0d1117;
        --table-border: #30363d;
        --table-row-border: #30363d;
        --table-alt-bg: #1c2128;
        --hr-bg: #30363d;
        --kbd-fg: #c9d1d9;
        --kbd-border: #6e7681;
        --mark-bg: rgba(210,153,34,0.4);
        --btn-border: #30363d;
        --hover-light-bg: rgba(56,139,253,0.15);
        --hover-purple-bg: rgba(188,140,255,0.15);
        --divider: #21262d;
        --selection-bg: rgba(88,166,255,0.28);
      `

// themeDarkTypography tunes glyph rendering for light-on-dark reading. Bright
// thin strokes bloom against a dark background (halation) and look heavier than
// they are, so we switch to grayscale antialiasing and open the tracking a
// hair. Applied for the explicit dark theme and, unless the reader forced
// light mode, for the system preference as well.
const themeDarkTypography = `
      :root[data-theme="dark"] body {
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        letter-spacing: 0.01em;
      }

      @media (prefers-color-scheme: dark) {
        :root:not([data-theme="light"]) body {
          -webkit-font-smoothing: antialiased;
          -moz-osx-font-smoothing: grayscale;
          letter-spacing: 0.01em;
        }
      }
      `

// themeSelection keeps text selection legible in both palettes.
const themeSelection = `
      ::selection {
        background: var(--selection-bg);
      }
      `

const themeToggleCSS = `
      .theme-toggle {
        position: fixed;
        top: 14px;
        right: 14px;
        z-index: 100;
        width: 36px;
        height: 36px;
        padding: 0;
        font-size: 17px;
        line-height: 1;
        color: var(--fg);
        background: var(--card-bg);
        border: 1px solid var(--btn-border);
        border-radius: 50%;
        cursor: pointer;
        box-shadow: var(--card-shadow);
      }

      .theme-toggle:hover {
        border-color: var(--link);
      }
      `

const themeButton = `<button id="theme-toggle" class="theme-toggle" type="button" aria-label="切换主题"></button>`

// themeHeadScript applies the saved theme before first paint to avoid flashing.
const themeHeadScript = `<script>
      (function () {
        var theme = null;
        try { theme = localStorage.getItem("markdir-theme"); } catch (e) {}
        if (theme === "light" || theme === "dark") {
          document.documentElement.setAttribute("data-theme", theme);
        }
      })();
    </script>`

// themeToggleScript cycles auto → light → dark and persists the choice.
// "auto" removes the attribute so the system-preference media query decides.
const themeToggleScript = `<script>
      (function () {
        var KEY = "markdir-theme";
        var MODES = ["auto", "light", "dark"];
        var ICONS = { auto: "🌗", light: "☀️", dark: "🌙" };
        var LABELS = { auto: "跟随系统", light: "浅色", dark: "深色" };
        var btn = document.getElementById("theme-toggle");
        if (!btn) return;
        var mq = window.matchMedia("(prefers-color-scheme: dark)");

        function mode() {
          var v = null;
          try { v = localStorage.getItem(KEY); } catch (e) {}
          return MODES.indexOf(v) >= 0 ? v : "auto";
        }

        function apply(m) {
          if (m === "auto") {
            document.documentElement.removeAttribute("data-theme");
          } else {
            document.documentElement.setAttribute("data-theme", m);
          }
          btn.textContent = ICONS[m];
          btn.title = LABELS[m] + "（点击切换）";
        }

        btn.addEventListener("click", function () {
          var next = MODES[(MODES.indexOf(mode()) + 1) % MODES.length];
          try { localStorage.setItem(KEY, next); } catch (e) {}
          apply(next);
        });

        function onSystemChange() {
          if (mode() === "auto") apply("auto");
        }
        if (mq.addEventListener) mq.addEventListener("change", onSystemChange);
        else if (mq.addListener) mq.addListener(onSystemChange);

        apply(mode());
      })();
    </script>`
