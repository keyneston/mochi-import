package gomochi

import (
	"net/http"
	"path"
)

type Card struct {
	ID             string               `json:"id,omitempty"`
	Content        string               `json:"content"`
	DeckID         string               `json:"deck-id"`
	TemplateID     string               `json:"template-id,omitempty"`
	Fields         map[string]CardField `json:"fields,omitempty"`
	Archived       bool                 `json:"archived?"`
	ReviewReversed bool                 `json:"review-reverse?"`
	Position       string               `json:"pos,omitempty"`
}

func (c *Card) SetTemplate(id string) {
	tInfo := templates.Get(id)
	if tInfo != nil {
		c.TemplateID = tInfo.ID
	} else {
		c.TemplateID = id
	}
}

func (c *Card) AddField(id, value string) {
	fieldInfo := templates.Get(c.TemplateID).Get(id)
	if fieldInfo != nil {
		id = fieldInfo.ID
	}
	if c.Fields == nil {
		c.Fields = make(map[string]CardField)
	}

	c.Fields[id] = CardField{id, value}
}

type CardField struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

func (c *Client) GetCard(id string) (*Card, error) {
	result := &Card{}

	if err := c.Request(path.Join(PathCard, sanitiseID(id)), http.MethodGet, nil, result); err != nil {
		return nil, err
	}

	return result, nil
}
