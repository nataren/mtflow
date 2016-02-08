package main

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"json"
	"log"
	"os"
	"path/filepath"
	"time"
)

type SearchEngine struct {

	//--- Public fields ---
	BaseStoragePath string
	IndexID         string

	//--- Private fields ---
	index bleve.Index
}

type SearchIndexingDocument struct {
	ID                      string
	MessageID               *int
	MessageFlowID           *string
	MessageSent             *Time
	MessageUserID           *string
	MessageEvent            *string
	MessageRawContent       *json.RawMessage
	MessageMessageID        *int
	MessageTags             *[]string
	MessageUUID             *string
	MessageExternalUserName *string
	MessageApp              *string
	MessageThreadId         *string
}

type SearchResult struct {
	Timestamp time.Time
	Summary   string
}

func exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (s *SearchEngine) Init() {

	// Validate index location
	indexPath := filepath.Join(s.BaseStoragePath, s.IndexID+".index")
	exists, err := exists(indexPath)
	if !exists || err != nil {
		log.Printf("search won't be available, index path not found: %s", indexPath)
		return
	}

	// Init the index
	indexMapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, indexMapping)
	if err != nil {
		log.Printf("search won't be available, could not create index: %s: %v", indexPath, err)
		return
	}

	// Have index available
	s.index = index
}

func (s *SearchEngine) Index(doc SearchIndexingDocument) {
	if err := s.index.Index(doc); err != nil {
		log.Println("could not index document: %v", doc)
	}
}

func (s *SearchEngine) Find(terms []string) []SearchResult {
	var results []SearchResult
	return results
}

func (s *SearchEngine) Format(searchResults []SearchResult) []string {
	var results []string
	for _, value := range searchResults {
		results = append(results, s.formatSearchResult(value))
	}
	return results
}

func (s *SearchEngine) formatSearchResult(result SearchResult) string {
	return fmt.Sprintf("[%v]: %v", result.Timestamp, result.Summary)
}
