package gomochi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

type Client struct {
	HTTPClient *http.Client `json:"-"`
	APIBase    string       `json:"api_base"`

	APIKey    string             `json:"-"`
	Templates *TemplateConfigSet `json:"templates"`
	Debug     bool               `json:"debug"`
	Noop      bool               `json:"noop"`
}

// Request makes the request against the mochi api.
//
// WARNING: This method is Private API, it may change at any point, and is only
// exposed as an escape hatch.
//
// - body: if included it will encode it as JSON and send it as the body of the request.
// - response: if set the response will be JSON decoded into the response
func (c *Client) Request(apiPath string, method string, body any, response any) error {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	apiBase := c.APIBase
	if apiBase == "" {
		apiBase = APIBase
	}

	u, err := url.Parse(apiBase)
	if err != nil {
		return fmt.Errorf("Error parsing %q: %w", apiBase, err)
	}
	u.Path = path.Join(u.Path, apiPath)

	encodedRequest := &bytes.Buffer{}
	if body != nil {
		if err := json.NewEncoder(encodedRequest).Encode(body); err != nil {
			return fmt.Errorf("Error encoding body: %w", err)
		}
	}

	req, err := http.NewRequest(method, u.String(), encodedRequest)
	if err != nil {
		return fmt.Errorf("Error building request: %w", err)
	}

	req.SetBasicAuth(c.APIKey, "")
	req.Header.Set("Content-Type", "application/json")

	if c.Noop {
		dump, _ := httputil.DumpRequest(req, true)
		log.Printf("%v", string(dump))
		return nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %w", err)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("Error decoding JSON response: %w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Error making request: status %d; message: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// sanitiseID removes characters such as '[]' which mochi likes to give you when you "copy id"
func sanitiseID(input string) string {
	return strings.Trim(input, "[]")
}
