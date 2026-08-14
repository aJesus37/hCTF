package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Category struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type Difficulty struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	TextColor string `json:"text_color"`
	SortOrder int    `json:"sort_order"`
}

var (
	catIDRe  = regexp.MustCompile(`id="cat-([^"]+)"`)
	diffIDRe = regexp.MustCompile(`id="diff-([^"]+)"`)
)

// parseSettingID extracts the entity id from the HTML row fragment returned by
// the settings create/update endpoints. Those endpoints render server-side
// HTML rows (`id="cat-..."` / `id="diff-..."`) via renderCategoryHTML and
// renderDifficultyHTML in internal/handlers/settings.go; the regexes above
// must stay in sync with that format so a server-side template change surfaces
// as an error instead of a silently blank id.
func parseSettingID(body []byte, re *regexp.Regexp) (string, error) {
	m := re.FindSubmatch(body)
	if m == nil {
		return "", fmt.Errorf("could not parse id from server response")
	}
	return string(m[1]), nil
}

func (c *Client) ListCategories() ([]Category, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/categories", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out []Category
	return out, decodeJSON(resp, &out)
}

func (c *Client) ListDifficulties() ([]Difficulty, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/difficulties", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out []Difficulty
	return out, decodeJSON(resp, &out)
}

func (c *Client) CreateCategory(name string, order int) (*Category, error) {
	form := url.Values{"name": {name}, "sort_order": {strconv.Itoa(order)}}
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/admin/categories", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	id, err := parseSettingID(body, catIDRe)
	if err != nil {
		return nil, err
	}
	return &Category{ID: id, Name: name}, nil
}

func (c *Client) UpdateCategory(id, name string, sortOrder int) (*Category, error) {
	form := url.Values{"name": {name}, "sort_order": {strconv.Itoa(sortOrder)}}
	req, _ := http.NewRequest("PUT", c.ServerURL+"/api/admin/categories/"+id, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	got, err := parseSettingID(body, catIDRe)
	if err != nil {
		return nil, err
	}
	return &Category{ID: got, Name: name, SortOrder: sortOrder}, nil
}

func (c *Client) DeleteCategory(id string) error {
	return c.doNoBody("DELETE", "/api/admin/categories/"+id)
}

func (c *Client) CreateDifficulty(name string, order int, color, textColor string) (*Difficulty, error) {
	form := url.Values{
		"name":       {name},
		"sort_order": {strconv.Itoa(order)},
		"color":      {color},
		"text_color": {textColor},
	}
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/admin/difficulties", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	id, err := parseSettingID(body, diffIDRe)
	if err != nil {
		return nil, err
	}
	return &Difficulty{ID: id, Name: name}, nil
}

func (c *Client) UpdateDifficulty(id, name, color, textColor string, sortOrder int) (*Difficulty, error) {
	form := url.Values{
		"name":       {name},
		"color":      {color},
		"text_color": {textColor},
		"sort_order": {strconv.Itoa(sortOrder)},
	}
	req, _ := http.NewRequest("PUT", c.ServerURL+"/api/admin/difficulties/"+id, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	got, err := parseSettingID(body, diffIDRe)
	if err != nil {
		return nil, err
	}
	return &Difficulty{ID: got, Name: name, Color: color, TextColor: textColor, SortOrder: sortOrder}, nil
}

func (c *Client) DeleteDifficulty(id string) error {
	return c.doNoBody("DELETE", "/api/admin/difficulties/"+id)
}
