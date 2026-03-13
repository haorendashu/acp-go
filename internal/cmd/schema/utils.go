package main

import "strings"

// commonAcronyms maps lowercase words to their Go-conventional uppercase form.
var commonAcronyms = map[string]string{
	"id": "ID", "url": "URL", "uri": "URI",
	"http": "HTTP", "https": "HTTPS", "json": "JSON",
	"api": "API", "sql": "SQL", "ssh": "SSH",
	"tcp": "TCP", "udp": "UDP", "ip": "IP",
	"html": "HTML", "css": "CSS", "xml": "XML",
	"rpc": "RPC", "tls": "TLS", "ssl": "SSL",
	"eof": "EOF", "sse": "SSE", "mcp": "MCP",
	"fs": "FS", "ui": "UI", "io": "IO",
}

// toMultiLineComment converts a multi-line string to Go comment format
func toMultiLineComment(s string) string {
	if s == "" {
		return ""
	}
	return "// " + strings.ReplaceAll(s, "\n", "\n// ") + "\n"
}

// toTitleCase converts snake_case, kebab-case, or camelCase string to TitleCase
// with Go-conventional acronym handling (e.g., ID, URL, HTTP).
func toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	var words []string

	// Handle snake_case, kebab-case, and space-separated
	if strings.ContainsAny(s, "_- ") {
		s = strings.ReplaceAll(s, "-", "_")
		s = strings.ReplaceAll(s, " ", "_")
		words = strings.Split(s, "_")
	} else {
		// Handle camelCase - split on uppercase letters
		var word strings.Builder
		for i, r := range s {
			if i > 0 && (r >= 'A' && r <= 'Z') {
				words = append(words, word.String())
				word.Reset()
			}
			word.WriteRune(r)
		}
		if word.Len() > 0 {
			words = append(words, word.String())
		}
	}

	// Capitalize each word, using Go-conventional acronyms
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		lower := strings.ToLower(word)
		if acronym, ok := commonAcronyms[lower]; ok {
			words[i] = acronym
		} else {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, "")
}

