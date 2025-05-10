package client

import (
	"agent/internal/models/req"
	"agent/internal/models/resp"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	http.Client
	Host   string
	Port   string
	Logger *zap.Logger
}

func (c *Client) GetTask() *resp.Task {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("http://%s:%s/internal/task", c.Host, c.Port)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.Logger.Error("error while creating request", zap.Error(err))
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := c.Do(req.WithContext(ctx))
	if err != nil {
		c.Logger.Error("error while sending request", zap.Error(err))

		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.Logger.Error("error while sending request", zap.String("status", response.Status))

		return nil
	}

	answer := struct {
		Task resp.Task `json:"task"`
	}{}

	err = json.NewDecoder(response.Body).Decode(&answer)
	if err != nil {
		return nil
	}

	return &answer.Task
}

func (c *Client) SendResult(result req.Result) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(result)
	if err != nil {
		c.Logger.Error("error while encoding result", zap.Error(err))
		return
	}

	url := fmt.Sprintf("http://%s:%s/internal/task", c.Host, c.Port)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		c.Logger.Error("error while creating request", zap.Error(err))
		return
	}

	response, err := c.Do(req.WithContext(ctx))
	if err != nil {
		c.Logger.Error("error while sending result", zap.Error(err))
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		c.Logger.Error("error while sending result", zap.String("status", response.Status))
	}
}
