package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/keyneston/mochi-import/gomochi"
)

const (
	APICard = "/cards"
)

var (
	templateID         string
	templateConfigPath string
	file               string
	deckID             string
	apiKey             string

	noop  bool
	debug bool
)

func main() {
	flag.StringVar(&apiKey, "key", os.Getenv("MOCHI_API_KEY"), "API Key to use for access")
	flag.StringVar(&file, "file", "", "CSV File to parse and import")
	flag.StringVar(&deckID, "deck", "", "DeckID to import to")
	flag.StringVar(&templateID, "template", "", "Template ID or Name to import as")
	flag.StringVar(&templateConfigPath, "template-config", "", "Path for template-config.json to import")
	flag.BoolVar(&debug, "debug", false, "Debug logging")
	flag.BoolVar(&noop, "noop", false, "Noop")
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

	if templateConfigPath != "" {
		if err := gomochi.LoadTemplateConfig(templateConfigPath); err != nil {
			log.Fatalf("Error loading templates from %q: %v", templateConfigPath, err)
		}
	}

	// card, err := c.GetCard("EQWk0qHZ")
	// prettyPrint(card)
	// log.Fatalf("err: %v", err)

	if err := importFile(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func prettyPrint(input any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(input)
}

func importFile() error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	c := csv.NewReader(f)

	headers, err := c.Read()
	switch {
	case err == io.EOF:
		return fmt.Errorf("Got EOF while reading headers")
	case err != nil:
		return err
	}

	client := gomochi.Client{
		APIKey: apiKey,
		Noop:   noop,
		Debug:  debug,
	}

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
	}

	return nil
}

func uploadCard(client gomochi.Client, headers, record []string) error {
	card := gomochi.Card{
		DeckID:     deckID,
		TemplateID: templateID,
	}

	if len(headers) != len(record) {
		log.Printf("Mismatching record numbers: header(%d) vs record(%v); Record: %#v", len(headers), len(record), record)
		return nil
	}

	for i := range headers {
		card.AddField(headers[i], record[i])
	}

	if err := client.Request(APICard, http.MethodPost, card, nil); err != nil {
		return err
	}

	return nil
}
