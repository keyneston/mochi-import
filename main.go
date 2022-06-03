package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	APIBase = "https://app.mochi.cards/api/"
	APICard = "/cards"
)

var (
	templateID string
	file       string
	deckID     string
	apiKey     string
	content    string
)

func main() {
	flag.StringVar(&apiKey, "key", os.Getenv("MOCHI_API_KEY"), "API Key to use for access")
	flag.StringVar(&file, "f", "", "CSV File to parse and import")
	flag.StringVar(&deckID, "d", "", "DeckID to import to")
	flag.StringVar(&templateID, "t", "", "Template ID to import as")
	flag.StringVar(&content, "content", "", "Content to set")
	flag.Parse()

	if apiKey == "" {
		log.Fatalf("Must set apiKey")
	}
	if deckID == "" {
		log.Fatalf("Must set deckID")
	}
	if file == "" {
		log.Fatalf("Must set file to import from")
	}

	if err := importFile(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func importFile() error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	c := csv.NewReader(f)

	client := http.DefaultClient

	headers, err := c.Read()
	switch {
	case err == io.EOF:
		return fmt.Errorf("Got EOF while reading headers")
	case err != nil:
		return err
	}

	count := 0
	for {
		record, err := c.Read()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		}

		if err := uploadCard(client, headers, record); err != nil {
			return err
		}

		count++
		if count == 10 {
			return nil
		}
	}

	return nil
}

type NewCardRequest struct {
	Content    string               `json:"content"`
	DeckID     string               `json:"deck-id"`
	TemplateID string               `json:"template-id,omitempty"`
	Fields     map[string]CardField `json:"fields,omitempty"`
}

type CardField struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

func uploadCard(client *http.Client, headers, record []string) error {
	cardRequest := NewCardRequest{
		Content:    content,
		DeckID:     deckID,
		TemplateID: templateID,
	}

	if len(headers) != len(record) {
		log.Printf("Mismatching record numbers: header(%d) vs record(%v); Record: %#v", len(headers), len(record), record)
		return nil
	}

	for i := range headers {
		if cardRequest.Fields == nil {
			cardRequest.Fields = make(map[string]CardField, len(headers))
		}

		cardRequest.Fields[headers[i]] = CardField{headers[i], record[i]}
	}

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(cardRequest); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, APIBase+APICard, buffer)
	if err != nil {
		return err
	}
	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		enc.Encode(cardRequest)
		return fmt.Errorf("Error setting row %v: status code %d; %v", record, resp.StatusCode, string(body))
	}

	return nil
}
