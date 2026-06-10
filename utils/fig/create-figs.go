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
	imps := []string{}
	figs := []string{}
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

			imps = append(imps, fmt.Sprintf("import %s from \"./images/%s\";\n", camelCase, name))
			figs = append(figs, fmt.Sprintf("<Fig src={%s} caption=\"changeme!\" />\n", camelCase))
		}
	}
	fmt.Print(strings.Join(imps, ""))
	fmt.Println()
	fmt.Print(strings.Join(figs, ""))
}
