/*
Credit to Gemini. Code used to generate the table.
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

// SearchResult models the required fields from the NDJSON search response
type SearchResult struct {
	ID string `json:"id"`
}

// Document models the response from the document details endpoint
type Document struct {
	Instances []Instance `json:"instances"`
}

type Src struct {
	Timestamp string `json:"ts"`
}

// Instance models the individual instance inside a document
type Instance struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Revision string   `json:"revision"`
	Links    []string `json:"links"`
	Sources  []Src    `json:"sources"`
}

func main() {
	searchURL := "https://dochunt.kmsec.uk/?q=links%3A%28http*+OR+https*%29&s=kmsec.uk&format=json"

	// 1. Fetch NDJSON search results
	resp, err := http.Get(searchURL)
	if err != nil {
		log.Fatalf("Failed to query search API: %v", err)
	}
	defer resp.Body.Close()

	// Use a map to keep track of distinct IDs
	distinctIDs := make(map[string]bool)
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		var sr SearchResult
		if err := json.Unmarshal(scanner.Bytes(), &sr); err != nil {
			log.Printf("Warning: Failed to parse NDJSON line: %v", err)
			continue
		}
		if sr.ID != "" {
			distinctIDs[sr.ID] = true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading NDJSON stream: %v", err)
	}

	// 2. Collect all instances
	var allInstances []Instance

	for id := range distinctIDs {
		docURL := fmt.Sprintf("https://dochunt.kmsec.uk/d/%s?format=json", id)

		docResp, err := http.Get(docURL)
		if err != nil {
			log.Printf("Failed to fetch document %s: %v", id, err)
			continue
		}

		var doc Document
		if err := json.NewDecoder(docResp.Body).Decode(&doc); err != nil {
			log.Printf("Failed to decode document %s: %v", id, err)
			docResp.Body.Close()
			continue
		}
		docResp.Body.Close()

		allInstances = append(allInstances, doc.Instances...)
	}

	// 3. Sort chronologically by the 'Created' field
	sort.Slice(allInstances, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, allInstances[i].Sources[0].Timestamp)
		t2, _ := time.Parse(time.RFC3339, allInstances[j].Sources[0].Timestamp)
		return t1.Before(t2)
	})

	// 4. Output Markdown Table
	fmt.Println("| time | id | revision | title | comma separated `link` |")
	fmt.Println("|---|---|---|---|---|")

	for _, inst := range allInstances {
		// Join links and escape any pipes in titles to prevent breaking the Markdown table
		formatted := make([]string, len(inst.Links))
		for i, link := range inst.Links {
			defanged := strings.Replace(link, "https://", "hxxps[://]", 1)
			defanged = strings.ReplaceAll(defanged, ".", "[.]")
			formatted[i] = fmt.Sprintf("`%s`", defanged)
		}

		joinedLinks := strings.Join(formatted, "<br/>")
		safeTitle := strings.ReplaceAll(inst.Title, "|", "&#124;")

		fmt.Printf("| %s | %s | %s | %s | %s |\n",
			inst.Sources[0].Timestamp,
			inst.ID,
			inst.Revision,
			safeTitle,
			joinedLinks,
		)
	}
}
