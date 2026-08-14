package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CustomCode holds the site-wide custom HTML injection and MOTD settings.
type CustomCode struct {
	HeadHTML    string `json:"head_html"`
	BodyEndHTML string `json:"body_end_html"`
	PagesJSON   string `json:"pages_json"`
	Motd        string `json:"motd"`
}

// GetCustomCode fetches the current custom HTML injection and MOTD settings.
func (c *Client) GetCustomCode() (*CustomCode, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/admin/custom-code", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out CustomCode
	return &out, decodeJSON(resp, &out)
}

// UpdateCustomCode sends only the provided fields. The server applies each
// field only when present in the form body (partial semantics), so has* flags
// indicate which fields were explicitly provided.
func (c *Client) UpdateCustomCode(head, body, pages, motd string, hasHead, hasBody, hasPages, hasMotd bool) error {
	vals := url.Values{}
	if hasHead {
		vals.Set("head_html", head)
	}
	if hasBody {
		vals.Set("body_end_html", body)
	}
	if hasPages {
		vals.Set("pages_json", pages)
	}
	if hasMotd {
		vals.Set("motd", motd)
	}
	req, _ := http.NewRequest("PUT", c.ServerURL+"/api/admin/custom-code", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}
