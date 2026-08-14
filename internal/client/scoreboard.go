package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ScoreboardEntry struct {
	Rank       int       `json:"rank"`
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	TeamID     *string   `json:"team_id,omitempty"`
	TeamName   *string   `json:"team_name,omitempty"`
	Points     int       `json:"points"`
	SolveCount int       `json:"solve_count"`
	LastSolve  time.Time `json:"last_solve"`
}

func (c *Client) SetScoreboardFreeze(frozen bool) error {
	val := "0"
	if frozen {
		val = "1"
	}
	body := strings.NewReader(url.Values{"freeze_enabled": {val}}.Encode())
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/admin/settings/freeze", body)
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

func (c *Client) GetScoreboard() ([]ScoreboardEntry, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/scoreboard", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out []ScoreboardEntry
	return out, decodeJSON(resp, &out)
}

func (c *Client) GetTeamScoreboard() ([]ScoreboardEntry, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/teams/scoreboard", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out []ScoreboardEntry
	return out, decodeJSON(resp, &out)
}

func (c *Client) GetScoreEvolution(mode string, limit int) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/scoreboard/evolution?mode=%s&limit=%d", c.ServerURL, url.QueryEscape(mode), limit), nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	return out, decodeJSON(resp, &out)
}

func (c *Client) SetAdminVisibility(visible bool) error {
	val := "0"
	if visible {
		val = "1"
	}
	body := strings.NewReader(url.Values{"admin_visible": {val}}.Encode())
	req, _ := http.NewRequest("POST", c.ServerURL+"/api/admin/settings/admin-visibility", body)
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

func (c *Client) ForceScoreRecord() error {
	return c.doNoBody("POST", "/api/admin/scoreboard/force-record")
}

func (c *Client) GetCTFtime() ([]byte, error) {
	req, _ := http.NewRequest("GET", c.ServerURL+"/api/ctftime", nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
