package main

const imageExportHead = `<script src="https://cdn.jsdelivr.net/npm/html-to-image@1.11.11/dist/html-to-image.js"></script>`

const imageExportCSS = `
      .image-export {
        position: fixed;
        top: 14px;
        right: 60px;
        z-index: 100;
        min-height: 36px;
        padding: 0 12px;
        color: var(--fg);
        background: var(--card-bg);
        border: 1px solid var(--btn-border);
        border-radius: 6px;
        cursor: pointer;
        box-shadow: var(--card-shadow);
      }

      .image-export:hover {
        border-color: var(--link);
      }

      .image-export:focus-visible {
        outline: 2px solid var(--link);
        outline-offset: 2px;
      }

      .image-export:disabled {
        cursor: wait;
        opacity: 0.7;
      }

      @media (max-width: 600px) {
        .image-export {
          top: 10px;
          right: 54px;
          min-height: 34px;
          padding: 0 10px;
          font-size: 13px;
        }
      }
`

const imageExportButton = `<button id="image-export" class="image-export" type="button">下载 PNG</button>`

const imageExportScript = `<script>
      (function () {
        var button = document.getElementById("image-export");
        var target = document.querySelector(".markdown-body");
        if (!button || !target) return;

        function waitForImages() {
          var images = Array.prototype.slice.call(target.querySelectorAll("img"));
          return Promise.all(images.map(function (image) {
            if (image.complete) return Promise.resolve();
            return new Promise(function (resolve) {
              image.addEventListener("load", resolve, { once: true });
              image.addEventListener("error", resolve, { once: true });
            });
          }));
        }

        function fileName() {
          var path = decodeURIComponent(window.location.pathname);
          var name = path.split("/").pop() || "markdown";
          return name.replace(/\.md$/i, "") + ".png";
        }

        async function exportImage() {
          if (!window.htmlToImage) {
            throw new Error("PNG 导出组件加载失败，请检查网络连接");
          }
          await Promise.all([document.fonts.ready, waitForImages()]);
          var dataUrl = await window.htmlToImage.toPng(target, {
            cacheBust: true,
            pixelRatio: 2,
            skipAutoScale: true,
            style: { margin: "0" },
          });
          var link = document.createElement("a");
          link.download = fileName();
          link.href = dataUrl;
          link.click();
        }

        button.addEventListener("click", async function () {
          button.disabled = true;
          button.textContent = "生成中…";
          try {
            await exportImage();
          } catch (error) {
            console.error("PNG export failed:", error);
            window.alert(error.message || "PNG 导出失败，请稍后重试");
          } finally {
            button.disabled = false;
            button.textContent = "下载 PNG";
          }
        });
      })();
    </script>`
