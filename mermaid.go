package main

// Mermaid support turns ```mermaid fenced code blocks into SVG diagrams in the
// browser. The library is fetched from a CDN only on pages that actually
// contain diagrams, rendering degrades to the original code block when the CDN
// is unreachable, and diagrams are re-rendered when the page theme changes.

const mermaidCSS = `
      .markdown-body .mermaid-diagram {
        margin: 0 0 20px;
        padding: 4px 0;
        text-align: center;
      }

      .markdown-body .mermaid-diagram[hidden] {
        display: none;
      }

      .markdown-body .mermaid-diagram svg {
        max-width: 100%;
        height: auto;
      }
`

const mermaidScript = `<script>
  (function () {
    // 库随二进制内嵌，由 /-/mermaid.min.js 提供；只有页面含图时才注入
    var LIB_URL = "/-/mermaid.min.js?v=` + mermaidVersion + `";
    var blocks = Array.prototype.slice.call(
      document.querySelectorAll(".markdown-body pre > code.language-mermaid")
    );
    if (!blocks.length) return;

    var items = blocks.map(function (code) {
      var pre = code.parentElement;
      var holder = document.createElement("div");
      holder.className = "mermaid-diagram";
      holder.hidden = true;
      pre.parentNode.insertBefore(holder, pre);
      return { pre: pre, holder: holder, source: code.textContent, rendered: false };
    });

    var darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
    var api = null;
    var renderSeq = 0;
    var queue = Promise.resolve();
    var currentTheme = "";
    var wantedTheme = "";

    // 主题跟随 data-theme 或系统偏好，与主题切换按钮的逻辑一致
    function effectiveTheme() {
      var forced = document.documentElement.getAttribute("data-theme");
      if (forced === "light" || forced === "dark") return forced;
      return darkQuery.matches ? "dark" : "light";
    }

    function cssVar(name, fallback) {
      var value = getComputedStyle(document.body).getPropertyValue(name).trim();
      return value || fallback;
    }

    // 图表专用的细线、圆角与配色，覆盖 mermaid 自动生成的内联样式；
    // 颜色在渲染时解析成字面值，避免 PNG 导出时 CSS 变量丢失
    function diagramCSS() {
      var line = cssVar("--diagram-line", "#a5adb8");
      var border = cssVar("--diagram-node-border", "#d8dee4");
      var muted = cssVar("--muted", "#57606a");
      return [
        ".node rect, .node circle, .node ellipse { rx: 7px; ry: 7px; }",
        ".edgePaths .path, .flowchart-link { stroke-width: 1.3px; stroke-linecap: round; }",
        ".edgeLabel text { fill: " + muted + "; }",
        ".edgeLabel rect { opacity: 1; rx: 4px; ry: 4px; }",
        ".marker path, .arrowheadPath { fill: " + line + "; stroke: " + line + "; }",
        ".cluster rect { rx: 8px; ry: 8px; stroke: " + border + "; }",
      ].join("");
    }

    // 用页面配色变量构造 mermaid 主题，明暗两套主题都能贴合卡片背景
    function initialize() {
      api.initialize({
        startOnLoad: false,
        securityLevel: "strict",
        suppressErrorRendering: true,
        theme: "base",
        themeCSS: diagramCSS(),
        darkMode: currentTheme === "dark",
        fontFamily: getComputedStyle(document.body).fontFamily,
        themeVariables: {
          fontSize: "15px",
          background: cssVar("--card-bg", "#ffffff"),
          mainBkg: cssVar("--diagram-node-bg", "#ffffff"),
          primaryColor: cssVar("--diagram-node-bg", "#ffffff"),
          primaryTextColor: cssVar("--fg", "#24292e"),
          primaryBorderColor: cssVar("--diagram-node-border", "#d8dee4"),
          nodeBorder: cssVar("--diagram-node-border", "#d8dee4"),
          nodeTextColor: cssVar("--fg", "#24292e"),
          secondaryColor: cssVar("--table-alt-bg", "#f6f8fa"),
          tertiaryColor: cssVar("--card-bg", "#ffffff"),
          lineColor: cssVar("--diagram-line", "#a5adb8"),
          defaultLinkColor: cssVar("--diagram-line", "#a5adb8"),
          textColor: cssVar("--fg", "#24292e"),
          titleColor: cssVar("--fg", "#24292e"),
          edgeLabelBackground: cssVar("--card-bg", "#ffffff"),
          clusterBkg: "transparent",
          clusterBorder: cssVar("--diagram-node-border", "#d8dee4"),
          // 时序图 / 状态图等继续沿用同一套中性色，避免默认的高饱和色块
          actorBkg: cssVar("--diagram-node-bg", "#ffffff"),
          actorBorder: cssVar("--diagram-node-border", "#d8dee4"),
          actorTextColor: cssVar("--fg", "#24292e"),
          actorLineColor: cssVar("--diagram-line", "#a5adb8"),
          signalColor: cssVar("--diagram-line", "#a5adb8"),
          signalTextColor: cssVar("--muted", "#57606a"),
          labelBoxBkgColor: cssVar("--card-bg", "#ffffff"),
          labelBoxBorderColor: cssVar("--diagram-node-border", "#d8dee4"),
          labelTextColor: cssVar("--fg", "#24292e"),
          loopTextColor: cssVar("--fg", "#24292e"),
          activationBkgColor: cssVar("--table-alt-bg", "#f6f8fa"),
          activationBorderColor: cssVar("--diagram-node-border", "#d8dee4"),
          noteBkgColor: cssVar("--table-alt-bg", "#f6f8fa"),
          noteBorderColor: cssVar("--diagram-node-border", "#d8dee4"),
          noteTextColor: cssVar("--fg", "#24292e"),
        },
        // htmlLabels 会生成 foreignObject，导出 PNG 时文字会丢失；顶层开关
        // 控制连线标签，flowchart 内的开关控制节点标签，两处都要关掉
        htmlLabels: false,
        flowchart: {
          htmlLabels: false,
          useMaxWidth: true,
          curve: "basis",
          nodeSpacing: 42,
          rankSpacing: 44,
          padding: 14,
        },
        sequence: { useMaxWidth: true },
        gantt: { useMaxWidth: true },
      });
    }

    function setFallback(item) {
      // 保留占位节点（隐藏），下次主题切换仍可重试渲染
      item.holder.innerHTML = "";
      item.holder.hidden = true;
      item.pre.hidden = false;
    }

    function renderItem(item) {
      var id = "markdir-mermaid-" + (++renderSeq);
      return api
        .render(id, item.source)
        .then(function (result) {
          item.holder.innerHTML = result.svg;
          item.holder.hidden = false;
          item.pre.hidden = true;
          item.rendered = true;
        })
        .catch(function (error) {
          console.error("mermaid render failed:", error);
          // 失败时 mermaid 会在页面里留下错误图，清掉并还原源码
          var leftover = document.getElementById("d" + id) || document.getElementById(id);
          if (leftover) leftover.remove();
          if (!item.rendered) setFallback(item);
        });
    }

    function renderAll(theme) {
      currentTheme = theme;
      initialize();
      return Promise.all(items.map(renderItem));
    }

    // 主题切换与系统偏好变化都排进同一条队列，避免并发渲染互相覆盖
    function scheduleRender() {
      if (!api) return;
      wantedTheme = effectiveTheme();
      queue = queue
        .then(function () {
          if (wantedTheme === currentTheme) return undefined;
          return renderAll(wantedTheme);
        })
        .catch(function (error) {
          console.error("mermaid re-render failed:", error);
        });
      window.markdirMermaidReady = queue;
    }

    function loadLibrary() {
      return new Promise(function (resolve, reject) {
        var script = document.createElement("script");
        script.addEventListener("load", function () {
          resolve(window.mermaid);
        });
        script.addEventListener("error", function () {
          reject(new Error("内嵌 mermaid 加载失败: " + LIB_URL));
        });
        script.async = true;
        script.src = LIB_URL;
        document.head.appendChild(script);
      });
    }

    // 首次加载也走这条队列，主题切换的渲染永远排在库就绪之后
    queue = loadLibrary()
      .then(function (loaded) {
        if (!loaded) throw new Error("mermaid 初始化失败");
        api = loaded;
        return renderAll(effectiveTheme());
      })
      .catch(function (error) {
        console.error("mermaid unavailable, keeping code blocks:", error);
        items.forEach(setFallback);
      });
    window.markdirMermaidReady = queue;

    new MutationObserver(scheduleRender).observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["data-theme"],
    });

    if (darkQuery.addEventListener) darkQuery.addEventListener("change", scheduleRender);
    else if (darkQuery.addListener) darkQuery.addListener(scheduleRender);
  })();
</script>`
