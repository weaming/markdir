package main

import (
	"regexp"
	"strings"
)

var styleBlockRe = regexp.MustCompile(`(?s)<style>(.*?)</style>`)

// minifyStyle 将每个 <style> 块内的空白折叠为单个空格，使渲染出的页面里
// CSS 保持单行。只在模板解析前调用，因此只压缩模板自带的样式，不影响用户内容。
func minifyStyle(html string) string {
	return styleBlockRe.ReplaceAllStringFunc(html, func(block string) string {
		css := block[len("<style>") : len(block)-len("</style>")]
		return "<style>" + strings.Join(strings.Fields(css), " ") + "</style>"
	})
}
