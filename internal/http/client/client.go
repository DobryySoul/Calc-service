package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DobryySoul/Calc-service/internal/result"
	"github.com/DobryySoul/Calc-service/internal/task"
)

type Client struct {
	http.Client
	Host string
	Port int
}

func (c *Client) GetTask() *task.Task {
	url := fmt.Sprintf("http://%s:%d/internal/task", c.Host, c.Port)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	answer := struct {
		Task task.Task `json:"task"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return nil
	}

	return &answer.Task
}

func (c *Client) SendResult(result result.Result) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(result)
	if err != nil {
		return
	}

	url := fmt.Sprintf("http://%s:%d/internal/task", c.Host, c.Port)

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return
	}

	defer resp.Body.Close()
}
