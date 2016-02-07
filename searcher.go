package main

import (
	"fmt"
	"time"
)

type Searcher struct {
}

type SearchResult struct {
	Timestamp time.Time
	Summary   string
}

func (s *Searcher) Find(terms []string) []SearchResult {
	var results []SearchResult
	return results
}

func (s *Searcher) Format(searchResults []SearchResult) []string {
	var results []string
	for _, value := range searchResults {
		results = append(results, s.formatSearchResult(value))
	}
	return results
}

func (s *Searcher) formatSearchResult(result SearchResult) string {
	return fmt.Sprintf("[%v]: %v", result.Timestamp, result.Summary)
}
