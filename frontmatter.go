package main

import (
	"html"
	"strconv"
	"strings"
)

// splitFrontMatter separates a leading YAML ("---") or TOML ("+++") front
// matter block from the markdown body. Its first non-empty, non-comment line
// must look like a key/value pair, so a document that merely starts with a
// horizontal rule is left untouched.
// The returned frontMatter excludes the delimiters; body is the remaining markdown.
func splitFrontMatter(src []byte) (frontMatter string, body []byte) {
	text := strings.TrimPrefix(string(src), "\ufeff")
	lines := strings.SplitAfter(text, "\n")
	if len(lines) < 3 {
		return "", src
	}

	var sep byte
	var isEnd func(string) bool
	switch trimEOL(lines[0]) {
	case "---":
		sep = ':'
		isEnd = func(s string) bool { return s == "---" || s == "..." }
	case "+++":
		sep = '='
		isEnd = func(s string) bool { return s == "+++" }
	default:
		return "", src
	}

	for i := 1; i < len(lines); i++ {
		if isEnd(trimEOL(lines[i])) {
			block := strings.Join(lines[1:i], "")
			if !looksLikeFrontMatter(block, sep) {
				return "", src
			}
			return block, []byte(strings.Join(lines[i+1:], ""))
		}
	}
	return "", src
}

// frontMatterValue returns the value of key in a YAML or TOML front matter
// block, or "" when absent.
func frontMatterValue(block, key string) string {
	for _, row := range parseFrontMatterRows(block) {
		if strings.EqualFold(row.key, key) {
			return row.value
		}
	}
	return ""
}

// frontMatterTable renders a front matter block as an HTML table the way GitHub
// does: one row per top-level key, the key in a <th> and its value in a <td>,
// with no header row. Returns "" when there are no keys.
func frontMatterTable(block string) string {
	rows := parseFrontMatterRows(block)
	if len(rows) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<table class=\"frontmatter\">\n<tbody>\n")
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

// parseFrontMatterRows dispatches to the YAML or TOML parser based on the
// block's key/value separator.
func parseFrontMatterRows(block string) []fmRow {
	if detectFrontMatterSep(block) == '=' {
		return parseTOMLRows(block)
	}
	return parseYAMLRows(block)
}

// parseYAMLRows splits a YAML front matter block into top-level "key: value"
// pairs. Indented lines (nested maps, lists) are folded into the current value,
// with list markers stripped, so multi-line metadata still reads sensibly.
func parseYAMLRows(yaml string) []fmRow {
	rows := []fmRow{}
	cur := -1

	for _, raw := range strings.Split(yaml, "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if line[0] != ' ' && line[0] != '\t' {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				if cur >= 0 {
					rows[cur].value = appendValue(rows[cur].value, trimmed)
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
			rows[cur].value = appendValue(rows[cur].value, trimmed)
		}
	}
	return rows
}

// parseTOMLRows splits a TOML front matter block into "key = value" pairs.
// Supports quoted strings, inline/multi-line arrays, comments, and [table]
// section headers.
func parseTOMLRows(toml string) []fmRow {
	rows := []fmRow{}
	section := ""
	lines := strings.Split(toml, "\n")

	for i := 0; i < len(lines); i++ {
		trimmed := stripTOMLComment(strings.TrimRight(lines[i], "\r"))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.TrimSpace(strings.Trim(trimmed, "[]"))
			continue
		}

		rawKey, rawVal, ok := strings.Cut(trimmed, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(rawKey)
		if section != "" {
			key = section + "." + key
		}

		val := strings.TrimSpace(rawVal)
		if strings.HasPrefix(val, "[") && !strings.HasSuffix(val, "]") {
			var arr strings.Builder
			arr.WriteString(val)
			for i+1 < len(lines) {
				i++
				next := stripTOMLComment(strings.TrimRight(lines[i], "\r"))
				if next == "" {
					continue
				}
				arr.WriteByte(' ')
				arr.WriteString(next)
				if strings.HasSuffix(next, "]") {
					break
				}
			}
			val = arr.String()
		}

		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			val = formatTOMLArray(val)
		} else {
			val = unquoteTOML(val)
		}

		rows = append(rows, fmRow{key: key, value: val})
	}
	return rows
}

// formatTOMLArray parses a TOML array literal like `["a", "b"]` into `"a, b"`.
func formatTOMLArray(s string) string {
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return ""
	}

	var items []string
	var cur strings.Builder
	inDouble, inSingle, escaped := false, false, false

	flush := func() {
		elem := strings.TrimSpace(cur.String())
		cur.Reset()
		if elem != "" {
			items = append(items, unquoteTOML(elem))
		}
	}

	for i := 0; i < len(inner); i++ {
		ch := inner[i]
		if escaped {
			cur.WriteByte(ch)
			escaped = false
			continue
		}
		if inDouble {
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inDouble = false
			}
			cur.WriteByte(ch)
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			cur.WriteByte(ch)
			continue
		}
		switch ch {
		case '"':
			inDouble = true
			cur.WriteByte(ch)
		case '\'':
			inSingle = true
			cur.WriteByte(ch)
		case ',':
			flush()
		default:
			cur.WriteByte(ch)
		}
	}
	flush()
	return strings.Join(items, ", ")
}

// stripTOMLComment removes a trailing unquoted # comment from a TOML line.
func stripTOMLComment(s string) string {
	inDouble, inSingle, escaped := false, false, false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if inDouble {
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inDouble = false
			}
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			continue
		}
		switch ch {
		case '"':
			inDouble = true
		case '\'':
			inSingle = true
		case '#':
			return strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
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

// unquoteTOML unquotes a TOML basic ("...") or literal ('...') string.
func unquoteTOML(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		if unquoted, err := strconv.Unquote(s); err == nil {
			return unquoted
		}
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// trimEOL strips a trailing newline (and carriage return) from a line.
func trimEOL(line string) string {
	return strings.TrimRight(line, "\r\n")
}

func isValidFrontMatterKey(key string) bool {
	return key != "" && !strings.ContainsAny(key, " \t:=")
}

// looksLikeFrontMatter reports whether the block's first non-empty, non-comment
// key/value line uses sep (':' for YAML, '=' for TOML).
func looksLikeFrontMatter(block string, sep byte) bool {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if sep == '=' && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			continue
		}
		key, _, ok := strings.Cut(trimmed, string(sep))
		return ok && isValidFrontMatterKey(strings.TrimSpace(key))
	}
	return false
}

// detectFrontMatterSep inspects the first key/value line to decide whether a
// front matter block uses YAML (':') or TOML ('=').
func detectFrontMatterSep(block string) byte {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			continue
		}
		colonIdx := strings.IndexByte(trimmed, ':')
		eqIdx := strings.IndexByte(trimmed, '=')
		if eqIdx > 0 && (colonIdx < 0 || eqIdx < colonIdx) && isValidFrontMatterKey(strings.TrimSpace(trimmed[:eqIdx])) {
			return '='
		}
		if colonIdx > 0 && (eqIdx < 0 || colonIdx < eqIdx) && isValidFrontMatterKey(strings.TrimSpace(trimmed[:colonIdx])) {
			return ':'
		}
		return ':'
	}
	return ':'
}
