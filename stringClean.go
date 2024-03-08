package main

import (
	"fmt"
	"regexp"
	"strings"
)

func UnixSafeFilename(input string) string {
	input = StripControl(input)
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ReplaceAll(input, "..", "_")
	input = strings.ReplaceAll(input, ".", "_")
	input = UnixPreFilter(input)
	input = strings.TrimPrefix(input, ".")
	input = strings.TrimPrefix(input, ".")
	input = strings.TrimSuffix(input, ".sh")
	input = TruncateString(input, 64)
	return input
}

func AlphaOnly(str string) string {
	alphafilter, _ := regexp.Compile("[^a-zA-Z]+")
	str = alphafilter.ReplaceAllString(str, "")
	return str
}

func NumOnly(str string) string {
	alphafilter, _ := regexp.Compile("[^0-9]+")
	str = alphafilter.ReplaceAllString(str, "")
	return str
}

func AlphaNumOnly(str string) string {
	alphafilter, _ := regexp.Compile("[^a-zA-Z0-9]+")
	str = alphafilter.ReplaceAllString(str, "")
	return str
}

func UnixPreFilter(str string) string {
	alphafilter, _ := regexp.Compile("[^a-zA-Z0-9-_]+")
	str = alphafilter.ReplaceAllString(str, "")
	return str
}

/* Shorten strings, end with elipsis */
func TruncateStringEllipsis(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}

/* Shorten strings */
func TruncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		bnoden = str[0:num]
	}
	return bnoden
}

/* Strip all but a-z */
func StripControlAndSpecial(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 32 && c < 127 {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

/* Sub specials with '?', sub newlines, returns and tabs with ' ' */
func SubControlAndSpecial(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := fmt.Sprintf("%c", i)
		if c[0] >= 32 && c[0] < 127 {
			b[bl] = c[0]
			bl++
		} else if c[0] == '\n' || c[0] == '\r' || c[0] == '\t' {
			b[bl] = ' '
			bl++
		} else {
			b[bl] = '?'
			bl++
		}
	}
	return string(b[:bl])
}

/* Strip lower ascii codes, sub newlines, returns and tabs with ' ' */
func StripControlAndSubSpecial(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c == '\n' || c == '\r' || c == '\t' {
			b[bl] = ' '
			bl++
		} else if c >= 32 && c != 127 {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

/* Strip lower ascii codes */
func StripControl(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := fmt.Sprintf("%c", i)
		if c[0] >= 32 && c[0] != 127 {
			b[bl] = c[0]
			bl++
		}
	}
	return string(b[:bl])
}
