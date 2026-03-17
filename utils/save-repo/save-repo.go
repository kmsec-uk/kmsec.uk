package main

/*
All credit to Gemini for whipping this up.
Used to capture GitHub repos from
contagious trader report
*/

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func main() {
	markdownFile := "index.mdx"
	content, err := os.ReadFile(markdownFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// regex to capture github urls
	re := regexp.MustCompile(`github\.com/([^/\s]+)/([^/\s#?)]+)`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	processed := make(map[string]bool)

	for _, match := range matches {
		org := match[1]
		repo := match[2]
		repoSlug := fmt.Sprintf("%s/%s", org, repo)

		if processed[repoSlug] {
			continue
		}

		fmt.Printf("--> Processing: %s\n", repoSlug)
		processRepo(org, repo)
		processed[repoSlug] = true
	}
}

func processRepo(org, repo string) {
	outDir := filepath.Join("output", org, repo)
	os.MkdirAll(outDir, 0755)

	// 1. Check API & Save Info
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s", org, repo)
	resp, err := http.Get(apiUrl)
	if err != nil || resp.StatusCode == 404 {
		fmt.Printf("   [!] Skip: %s (404 or Error)\n", repo)
		return
	}
	defer resp.Body.Close()

	infoFile, _ := os.Create(filepath.Join(outDir, "info.json"))
	io.Copy(infoFile, resp.Body)
	infoFile.Close()

	// 2. Clone Repository
	tmpClonePath := filepath.Join(outDir, "temp_clone")

	cmd := exec.Command("git", "clone", "--quiet", fmt.Sprintf("https://github.com/%s/%s", org, repo), tmpClonePath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("   [!] Clone failed: %v\n", err)
		return
	}

	// 3. Zip and Cleanup
	zipPath := filepath.Join(outDir, "repo.zip")
	if err := zipFolder(tmpClonePath, zipPath); err != nil {
		fmt.Printf("   [!] Zip failed: %v\n", err)
	} else {
		fmt.Printf("   [✓] Success: Saved to %s\n", outDir)
	}

	os.RemoveAll(tmpClonePath)
}

func zipFolder(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		header, _ := zip.FileInfoHeader(info)
		header.Name, _ = filepath.Rel(source, path)
		header.Method = zip.Deflate

		writer, _ := archive.CreateHeader(header)
		file, _ := os.Open(path)
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}
