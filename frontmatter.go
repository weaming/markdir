package main

import (
	"html"
	"strings"
)

// splitFrontMatter separates a leading YAML front matter block from the
// markdown body. A block starts with a "---" line and ends at the next "---"
// or "..." line; its first non-empty line must look like a "key: value" pair,
// so a document that merely starts with a horizontal rule is left untouched.
// The returned yaml excludes the delimiters; body is the remaining markdown.
func splitFrontMatter(src []byte) (yaml string, body []byte) {
	text := strings.TrimPrefix(string(src), "\ufeff")
	lines := strings.SplitAfter(text, "\n")
	if len(lines) < 3 || trimEOL(lines[0]) != "---" {
		return "", src
	}

	for i := 1; i < len(lines); i++ {
		if t := trimEOL(lines[i]); t == "---" || t == "..." {
			block := strings.Join(lines[1:i], "")
			if !looksLikeYAML(block) {
				return "", src
			}
			return block, []byte(strings.Join(lines[i+1:], ""))
		}
	}
	return "", src
}

// frontMatterValue returns the scalar value of key in a front matter block, or
// "" when absent. Only simple "key: value" lines are handled.
func frontMatterValue(yaml, key string) string {
	for _, line := range strings.Split(yaml, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(k), key) {
			continue
		}
		return strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return ""
}

// frontMatterTable renders a front matter block as an HTML table the way GitHub
// does: one row per top-level key, the key in a <th> and its value in a <td>,
// with no header row. Returns "" when there are no keys.
func frontMatterTable(yaml string) string {
	rows := parseFrontMatterRows(yaml)
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<table>\n<tbody>\n")
	for _, row := range rows {
		value := strings.ReplaceAll(html.EscapeString(row.value), "\n", "<br>")
		sb.WriteString("<tr>\n<th>")
		sb.WriteString(html.EscapeString(row.key))
		sb.WriteString("</th>\n<td>")
		sb.WriteString(value)
		sb.WriteString("</td>\n</tr>\n")
	}
	sb.WriteString("</tbody>\n</table>\n")
	return sb.String()
}

// fmRow is one key/value pair of a front matter block.
type fmRow struct {
	key   string
	value string
}

// parseFrontMatterRows splits a front matter block into top-level "key: value"
// pairs. Indented lines (nested maps, lists) are folded into the current value,
// with list markers stripped, so multi-line metadata still reads sensibly.
func parseFrontMatterRows(yaml string) []fmRow {
	rows := []fmRow{}
	cur := -1

	for _, raw := range strings.Split(yaml, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				if cur >= 0 {
					rows[cur].value = appendValue(rows[cur].value, strings.TrimSpace(line))
				}
				continue
			}
			rows = append(rows, fmRow{
				key:   strings.TrimSpace(key),
				value: unquote(strings.TrimSpace(value)),
			})
			cur = len(rows) - 1
			continue
		}
		if cur >= 0 {
			rows[cur].value = appendValue(rows[cur].value, strings.TrimSpace(line))
		}
	}
	return rows
}

// appendValue joins a continuation line onto an existing value, dropping a
// leading YAML list marker.
func appendValue(existing, add string) string {
	add = strings.TrimPrefix(add, "- ")
	if existing == "" {
		return add
	}
	return existing + "\n" + add
}

// unquote strips a matching pair of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// trimEOL strips a trailing newline (and carriage return) from a line.
func trimEOL(line string) string {
	return strings.TrimRight(line, "\r\n")
}

// looksLikeYAML reports whether the block's first non-empty line is a "key:" pair.
func looksLikeYAML(block string) bool {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, _, ok := strings.Cut(trimmed, ":")
		key = strings.TrimSpace(key)
		return ok && key != "" && !strings.ContainsAny(key, " \t")
	}
	return false
}
