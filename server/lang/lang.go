package lang

import (
	"fmt"
)

var langToExt = map[string]string{
	"bash":   "sh",
	"c":      "c",
	"golang": "go",
	"perl":   "pl",
	"python": "py",
	"ruby":   "rb",
}

func ToExt(lang string) (string, error) {
	if ext, ok := langToExt[lang]; ok {
		return ext, nil
	}
	return "", fmt.Errorf("Language not supported")
}
