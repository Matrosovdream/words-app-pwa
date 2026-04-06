package parser

import (
	"strings"
	"unicode"
)

// Tokenize splits text into lowercase word tokens. Keeps only alphabetic
// sequences of 3+ characters.
func Tokenize(text string) []string {
	words := make([]string, 0, 64)
	var b strings.Builder
	flush := func() {
		if b.Len() >= 3 {
			words = append(words, b.String())
		}
		b.Reset()
	}
	for _, r := range text {
		if unicode.IsLetter(r) || r == '\'' || r == '-' {
			b.WriteRune(unicode.ToLower(r))
		} else {
			flush()
		}
	}
	flush()
	return words
}

// SplitSentences returns rough sentence splits from a body of text.
func SplitSentences(text string) []string {
	text = strings.Join(strings.Fields(text), " ")
	out := make([]string, 0, 8)
	var b strings.Builder
	for _, r := range text {
		b.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			s := strings.TrimSpace(b.String())
			if len(s) > 0 {
				out = append(out, s)
			}
			b.Reset()
		}
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		out = append(out, s)
	}
	return out
}

// FindSampleSentence returns the first sentence containing the given lowercased word.
func FindSampleSentence(sentences []string, lemma string) string {
	needle := strings.ToLower(lemma)
	for _, s := range sentences {
		if strings.Contains(strings.ToLower(s), needle) {
			if len(s) > 300 {
				return s[:300] + "…"
			}
			return s
		}
	}
	return ""
}
