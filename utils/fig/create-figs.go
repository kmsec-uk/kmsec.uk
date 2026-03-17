package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	files, _ := os.ReadDir("./images/")

	for _, file := range files {
		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if ext == ".png" || ext == ".jpg" {
			base := strings.TrimSuffix(name, filepath.Ext(name))

			words := strings.FieldsFunc(base, func(r rune) bool {
				return r == '-' || r == '_' || r == ' '
			})

			// convert to CamelCase
			for i, word := range words {
				runes := []rune(word)
				runes[0] = unicode.ToUpper(runes[0])
				words[i] = string(runes)
			}
			camelCase := strings.Join(words, "")

			fmt.Printf("import %s from \"./images/%s\";\n", camelCase, name)
			fmt.Printf("<Fig src={%s} caption=\"changeme!\" />\n", camelCase)
		}
	}
}
