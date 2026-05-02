package codeparser

import (
	"regexp"
	"strings"
)

type Language struct {
	Name        string
	Confidence  float64
	Patterns    []*regexp.Regexp
	Shebangs    []string
	FileExts    []string
}

var languages = map[string]*Language{
	"python": {
		Name:       "python",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^def\s+\w+\s*\(`),
			regexp.MustCompile(`^class\s+\w+.*:`),
			regexp.MustCompile(`^import\s+\w+`),
			regexp.MustCompile(`^from\s+\w+\s+import`),
			regexp.MustCompile(`if\s+__name__\s*==\s*['"]__main__['"]`),
			regexp.MustCompile(`print\s*\(`),
		},
		Shebangs: []string{"#!/usr/bin/env python", "#!/usr/bin/python"},
		FileExts: []string{".py", ".pyw", ".pyi"},
	},
	"javascript": {
		Name:       "javascript",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^function\s+\w+\s*\(`),
			regexp.MustCompile(`^const\s+\w+\s*=\s*\(`),
			regexp.MustCompile(`^let\s+\w+\s*=\s*\(`),
			regexp.MustCompile(`^var\s+\w+\s*=\s*\(`),
			regexp.MustCompile(`^=>\s*\{`),
			regexp.MustCompile(`^export\s+(default\s+)?(function|const|let|var|class)`),
			regexp.MustCompile(`^import\s+.*\s+from\s+['"]`),
			regexp.MustCompile(`console\.log\s*\(`),
			regexp.MustCompile(`require\s*\(`),
		},
		Shebangs: []string{"#!/usr/bin/env node", "#!/bin/node"},
		FileExts: []string{".js", ".jsx", ".mjs", ".cjs"},
	},
	"typescript": {
		Name:       "typescript",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^function\s+\w+\s*<.*>\s*\(`),
			regexp.MustCompile(`^interface\s+\w+\s*\{`),
			regexp.MustCompile(`^type\s+\w+\s*=`),
			regexp.MustCompile(`:\s*(string|number|boolean|any|void|never)\s*[;=,)]`),
			regexp.MustCompile(`^import\s+.*\s+from\s+['"].*['"]\s*;?\s*$`),
		},
		Shebangs: []string{},
		FileExts: []string{".ts", ".tsx", ".mts", ".cts"},
	},
	"go": {
		Name:       "go",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^func\s+\w+\s*\(`),
			regexp.MustCompile(`^package\s+\w+`),
			regexp.MustCompile(`^import\s+\(`),
			regexp.MustCompile(`^type\s+\w+\s+struct\s*\{`),
			regexp.MustCompile(`^type\s+\w+\s+interface\s*\{`),
			regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+\w+\s*\(`),
		},
		Shebangs: []string{},
		FileExts: []string{".go"},
	},
	"rust": {
		Name:       "rust",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^fn\s+\w+\s*[<(]`),
			regexp.MustCompile(`^struct\s+\w+`),
			regexp.MustCompile(`^impl\s+(\w+\s+for\s+)?\w+`),
			regexp.MustCompile(`^enum\s+\w+`),
			regexp.MustCompile(`^use\s+\w+::`),
			regexp.MustCompile(`^mod\s+\w+`),
			regexp.MustCompile(`println!\s*\(`),
			regexp.MustCompile(`vec!\s*\(`),
		},
		Shebangs: []string{"#!/usr/bin/env rustc"},
		FileExts: []string{".rs"},
	},
	"ruby": {
		Name:       "ruby",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^def\s+\w+(\s*\(.*\))?`),
			regexp.MustCompile(`^class\s+\w+(\s*<\s*\w+)?`),
			regexp.MustCompile(`^module\s+\w+`),
			regexp.MustCompile(`^require\s+['"]`),
			regexp.MustCompile(`^puts\s+`),
			regexp.MustCompile(`end\s*$`),
		},
		Shebangs: []string{"#!/usr/bin/env ruby", "#!/usr/bin/ruby"},
		FileExts: []string{".rb", ".rake", ".gemspec"},
	},
	"java": {
		Name:       "java",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^public\s+(static\s+)?class\s+\w+`),
			regexp.MustCompile(`^private\s+(static\s+)?class\s+\w+`),
			regexp.MustCompile(`^class\s+\w+(\s+extends\s+\w+)?(\s+implements\s+\w+)?`),
			regexp.MustCompile(`^public\s+(static\s+)?void\s+main\s*\(`),
			regexp.MustCompile(`System\.out\.println\s*\(`),
		},
		Shebangs: []string{},
		FileExts: []string{".java"},
	},
	"kotlin": {
		Name:       "kotlin",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^fun\s+\w+`),
			regexp.MustCompile(`^class\s+\w+(\s*:\s*\w+)?`),
			regexp.MustCompile(`^data\s+class\s+\w+`),
			regexp.MustCompile(`^sealed\s+class\s+\w+`),
			regexp.MustCompile(`^object\s+\w+`),
			regexp.MustCompile(`println\s*\(`),
		},
		Shebangs: []string{},
		FileExts: []string{".kt", ".kts"},
	},
	"swift": {
		Name:       "swift",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^func\s+\w+`),
			regexp.MustCompile(`^struct\s+\w+`),
			regexp.MustCompile(`^class\s+\w+(\s*:\s*\w+)?`),
			regexp.MustCompile(`^enum\s+\w+`),
			regexp.MustCompile(`^import\s+\w+`),
			regexp.MustCompile(`^var\s+\w+:\s*\w+`),
			regexp.MustCompile(`^let\s+\w+:\s*\w+`),
		},
		Shebangs: []string{"#!/usr/bin/swift"},
		FileExts: []string{".swift"},
	},
	"cpp": {
		Name:       "cpp",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^#include\s*<.*>`),
			regexp.MustCompile(`^#include\s*"[^"]*"`),
			regexp.MustCompile(`^int\s+main\s*\(`),
			regexp.MustCompile(`^void\s+\w+\s*\(`),
			regexp.MustCompile(`^class\s+\w+(\s*:\s*\w+)?\s*\{`),
			regexp.MustCompile(`^namespace\s+\w+`),
			regexp.MustCompile(`std::\w+`),
		},
		Shebangs: []string{},
		FileExts: []string{".cpp", ".cc", ".cxx", ".hpp", ".h", ".hxx"},
	},
	"c": {
		Name:       "c",
		Patterns:   []*regexp.Regexp{
			regexp.MustCompile(`^#include\s*<stdio\.h>`),
			regexp.MustCompile(`^#include\s*<stdlib\.h>`),
			regexp.MustCompile(`^int\s+main\s*\(`),
			regexp.MustCompile(`^void\s+\w+\s*\(`),
			regexp.MustCompile(`^printf\s*\(`),
			regexp.MustCompile(`^typedef\s+struct`),
		},
		Shebangs: []string{},
		FileExts: []string{".c", ".h"},
	},
}

func DetectLanguage(code string) (string, float64) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "unknown", 0
	}

	lines := strings.Split(code, "\n")
	firstLine := strings.TrimSpace(lines[0])

	shebangMatch := regexp.MustCompile(`^#!/.*/(\w+)`)
	if matches := shebangMatch.FindStringSubmatch(firstLine); len(matches) > 1 {
		shebangLang := matches[1]
		switch shebangLang {
		case "python", "python3", "python2":
			return "python", 0.95
		case "node", "nodejs":
			return "javascript", 0.95
		case "ruby", "rbenv":
			return "ruby", 0.95
		case "swift":
			return "swift", 0.95
		case "rustc", "cargo":
			return "rust", 0.95
		case "go":
			return "go", 0.95
		}
	}

	var scores = make(map[string]float64)
	for lang, langInfo := range languages {
		scores[lang] = 0
		for _, pattern := range langInfo.Patterns {
			if pattern.MatchString(code) {
				scores[lang] += 1.0
			}
		}
		if len(langInfo.Shebangs) > 0 {
			for _, sb := range langInfo.Shebangs {
				if strings.HasPrefix(code, sb) {
					scores[lang] += 2.0
					break
				}
			}
		}
	}

	var bestLang string
	var bestScore float64
	totalPatterns := 0

	for lang, info := range languages {
		totalPatterns += len(info.Patterns)
		if scores[lang] > bestScore {
			bestScore = scores[lang]
			bestLang = lang
		}
	}

	if bestScore == 0 {
		return "unknown", 0
	}

	confidence := bestScore / float64(totalPatterns) * 100
	if confidence > 95 {
		confidence = 95
	}

	if strings.Contains(code, "interface ") && strings.Contains(code, ": ") {
		return "typescript", confidence
	}

	return bestLang, confidence
}

func IsValidLanguage(lang string) bool {
	_, ok := languages[lang]
	return ok
}

func GetSupportedLanguages() []string {
	var langs []string
	for lang := range languages {
		langs = append(langs, lang)
	}
	return langs
}