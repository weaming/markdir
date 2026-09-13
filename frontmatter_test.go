package main

import (
	"strings"
	"testing"
)

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
		"<table>", "<tbody>",
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
