package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkdownRendersRawHTML(t *testing.T) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(`<a id="toc_70"></a>`), &buf); err != nil {
		t.Fatalf("render markdown: %v", err)
	}

	if got := buf.String(); !strings.Contains(got, `<a id="toc_70"></a>`) {
		t.Fatalf("raw HTML anchor was not preserved: %s", got)
	}
}

func TestSplitFrontMatter(t *testing.T) {
	cases := []struct {
		name     string
		src      string
		wantYAML string
		wantBody string
	}{
		{
			name:     "plain",
			src:      "name: web-design\ndescription: a skill\n---\n# Title\n\nbody\n",
			wantYAML: "",
			wantBody: "name: web-design\ndescription: a skill\n---\n# Title\n\nbody\n",
		},
		{
			name:     "yaml block",
			src:      "---\nname: web-design\ndescription: a skill\n---\n# Title\n\nbody\n",
			wantYAML: "name: web-design\ndescription: a skill\n",
			wantBody: "# Title\n\nbody\n",
		},
		{
			name:     "ellipsis terminator",
			src:      "---\ntitle: Hi\n...\nbody\n",
			wantYAML: "title: Hi\n",
			wantBody: "body\n",
		},
		{
			name:     "unterminated is untouched",
			src:      "---\nname: web-design\n# Title\n",
			wantYAML: "",
			wantBody: "---\nname: web-design\n# Title\n",
		},
		{
			name:     "horizontal rules are not front matter",
			src:      "---\n\n---\n\ntext\n",
			wantYAML: "",
			wantBody: "---\n\n---\n\ntext\n",
		},
		{
			name:     "bom prefix",
			src:      "\ufeff---\ntitle: Hi\n---\nbody\n",
			wantYAML: "title: Hi\n",
			wantBody: "body\n",
		},
		{
			name:     "crlf",
			src:      "---\r\ntitle: Hi\r\n---\r\nbody\r\n",
			wantYAML: "title: Hi\r\n",
			wantBody: "body\r\n",
		},
		{
			name:     "toml block",
			src:      "+++\ntitle       = \"Anthropic（ANT）IPO\"\ndate        = 2026-11-20T02:26:05+08:00\ndraft       = false\n+++\n\nbody\n",
			wantYAML: "title       = \"Anthropic（ANT）IPO\"\ndate        = 2026-11-20T02:26:05+08:00\ndraft       = false\n",
			wantBody: "\nbody\n",
		},
		{
			name:     "toml bom and crlf",
			src:      "\ufeff+++\r\ntitle = \"Hi\"\r\n+++\r\nbody\r\n",
			wantYAML: "title = \"Hi\"\r\n",
			wantBody: "body\r\n",
		},
		{
			name:     "toml unterminated is untouched",
			src:      "+++\ntitle = \"Hi\"\n# Heading\n",
			wantYAML: "",
			wantBody: "+++\ntitle = \"Hi\"\n# Heading\n",
		},
		{
			name:     "empty plus delimiters are not front matter",
			src:      "+++\n\n+++\n\ntext\n",
			wantYAML: "",
			wantBody: "+++\n\n+++\n\ntext\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			yaml, body := splitFrontMatter([]byte(tc.src))
			if yaml != tc.wantYAML {
				t.Errorf("yaml = %q, want %q", yaml, tc.wantYAML)
			}
			if string(body) != tc.wantBody {
				t.Errorf("body = %q, want %q", body, tc.wantBody)
			}
		})
	}
}

func TestFrontMatterValue(t *testing.T) {
	yaml := "name: web-design\n\ntitle: \"Web 设计\"\nquoted: 'x'\n"
	cases := map[string]string{
		"name":   "web-design",
		"title":  "Web 设计",
		"quoted": "x",
		"absent": "",
	}
	for key, want := range cases {
		if got := frontMatterValue(yaml, key); got != want {
			t.Errorf("frontMatterValue(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestFrontMatterTable(t *testing.T) {
	if got := frontMatterTable(""); got != "" {
		t.Errorf("empty front matter should render nothing, got %q", got)
	}

	table := frontMatterTable("name: web-design\ndescription: a skill\n")
	for _, want := range []string{
		`<table class="frontmatter">`, "<tbody>",
		"<th>name</th>", "<td>web-design</td>",
		"<th>description</th>", "<td>a skill</td>",
	} {
		if !strings.Contains(table, want) {
			t.Errorf("table missing %q\n got: %s", want, table)
		}
	}
	if strings.Contains(table, "<thead>") {
		t.Error("table should have no header row (matching GitHub)")
	}
}

func TestFrontMatterTableEscapesAndFolds(t *testing.T) {
	table := frontMatterTable("title: \"<b>hi</b>\"\ntags:\n  - a\n  - b\n")

	if strings.Contains(table, "<b>hi</b>") {
		t.Error("value HTML should be escaped")
	}
	if !strings.Contains(table, "&lt;b&gt;hi&lt;/b&gt;") {
		t.Errorf("expected escaped value, got: %s", table)
	}
	if !strings.Contains(table, "<td>a<br>b</td>") {
		t.Errorf("list items should fold into one <br>-joined cell, got: %s", table)
	}
}

func TestTOMLFrontMatterValueAndTable(t *testing.T) {
	src := strings.Join([]string{
		"+++",
		`title       = "Anthropic（ANT）IPO 股权图解：谁持股，谁投票，谁选董事？"`,
		`date        = 2026-11-20T02:26:05+08:00`,
		`author      = "weaming"`,
		`description = ""`,
		`tags        = ["AI", "MCP", "工具, 开源"]`,
		`categories  = []`,
		`draft       = false # inline comment`,
		`toc         = true`,
		`multi       = [`,
		`  "one",`,
		`  "two",`,
		`]`,
		`[params]`,
		`note        = "<b>safe</b>"`,
		"+++",
		"",
		"正文内容",
	}, "\n")

	fm, body := splitFrontMatter([]byte(src))
	if fm == "" {
		t.Fatal("expected TOML front matter to be parsed")
	}
	if strings.TrimSpace(string(body)) != "正文内容" {
		t.Fatalf("unexpected body: %q", string(body))
	}

	wantValues := map[string]string{
		"title":       "Anthropic（ANT）IPO 股权图解：谁持股，谁投票，谁选董事？",
		"date":        "2026-11-20T02:26:05+08:00",
		"author":      "weaming",
		"description": "",
		"tags":        "AI, MCP, 工具, 开源",
		"categories":  "",
		"draft":       "false",
		"toc":         "true",
		"multi":       "one, two",
		"params.note": "<b>safe</b>",
		"absent":      "",
	}
	for k, want := range wantValues {
		if got := frontMatterValue(fm, k); got != want {
			t.Errorf("frontMatterValue(%q) = %q, want %q", k, got, want)
		}
	}

	table := frontMatterTable(fm)
	for _, want := range []string{
		"<th>title</th>\n<td>Anthropic（ANT）IPO 股权图解：谁持股，谁投票，谁选董事？</td>",
		"<th>date</th>\n<td>2026-11-20T02:26:05+08:00</td>",
		"<th>tags</th>\n<td>AI, MCP, 工具, 开源</td>",
		"<th>categories</th>\n<td></td>",
		"<th>draft</th>\n<td>false</td>",
		"<th>multi</th>\n<td>one, two</td>",
		"<th>params.note</th>\n<td>&lt;b&gt;safe&lt;/b&gt;</td>",
	} {
		if !strings.Contains(table, want) {
			t.Errorf("table missing %q\ngot:\n%s", want, table)
		}
	}
}
