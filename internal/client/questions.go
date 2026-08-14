package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type QuestionDetail struct {
	ID            string  `json:"id"`
	ChallengeID   string  `json:"challenge_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Points        int     `json:"points"`
	FlagMask      *string `json:"flag_mask"`
	CaseSensitive bool    `json:"case_sensitive"`
}

func (c *Client) ListQuestions(challengeID string) ([]Question, error) {
	_, qs, err := c.GetChallengeWithQuestions(challengeID)
	return qs, err
}

func (c *Client) CreateQuestion(challengeID, name, description, flag string, flagMask *string, caseSensitive bool, points int) (*QuestionDetail, error) {
	body, _ := json.Marshal(map[string]any{
		"challenge_id":   challengeID,
		"name":           name,
		"description":    description,
		"flag":           flag,
		"flag_mask":      flagMask,
		"case_sensitive": caseSensitive,
		"points":         points,
	})
	var out QuestionDetail
	return &out, c.doJSON("POST", "/api/admin/questions", bytes.NewReader(body), &out)
}

func (c *Client) GetQuestion(id string) (*QuestionDetail, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/admin/questions/%s", c.ServerURL, id), nil)
	req.Header.Set("Accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var out QuestionDetail
	return &out, decodeJSON(resp, &out)
}

type QuestionUpdate struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	Flag          *string `json:"flag,omitempty"`
	FlagMask      *string `json:"flag_mask,omitempty"`
	Points        *int    `json:"points,omitempty"`
	CaseSensitive *bool   `json:"case_sensitive,omitempty"`
}

func (c *Client) UpdateQuestion(id string, u QuestionUpdate) error {
	body, _ := json.Marshal(u)
	return c.doStatus("PUT", "/api/admin/questions/"+id, bytes.NewReader(body))
}

func (c *Client) DeleteQuestion(id string) error {
	return c.doNoBody("DELETE", "/api/admin/questions/"+id)
}
