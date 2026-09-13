package main

import (
	"slices"
	"testing"
)

func TestParseIgnore(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "defaults", raw: "", want: []string{".git/", ".DS_Store"}},
		{name: "extra dir", raw: " node_modules ", want: []string{".git/", ".DS_Store", "node_modules"}},
		{name: "skip empty", raw: "a,,b", want: []string{".git/", ".DS_Store", "a", "b"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseIgnore(c.raw); !slices.Equal(got, c.want) {
				t.Fatalf("parseIgnore(%q) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}

func TestIsIgnored(t *testing.T) {
	list := parseIgnore("node_modules, tmp/")

	cases := []struct {
		name string
		want bool
	}{
		{name: ".DS_Store", want: true},     // macOS 元数据文件默认隐藏
		{name: ".git/", want: true},         // 默认隐藏的目录
		{name: ".github/", want: false},     // 前缀相同的目录不受影响
		{name: "node_modules/", want: true}, // 额外指定的目录
		{name: "node_modules", want: true},  // 同名文件
		{name: "tmp/", want: true},          // 带 / 的条目匹配目录
		{name: "tmp", want: false},          // 带 / 的条目不匹配文件
		{name: "main.go", want: false},
		{name: "readme.md", want: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isIgnored(c.name, list); got != c.want {
				t.Fatalf("isIgnored(%q) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}
