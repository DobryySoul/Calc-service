package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/DobryySoul/Calc-service/internal/http/models/req"
	"github.com/DobryySoul/Calc-service/internal/http/models/resp"
	"github.com/DobryySoul/Calc-service/pkg/calculation"
)

type Client struct {
	http.Client
	Host string
	Port int
}

func (client *Client) GetTask() *resp.Task {
	url := fmt.Sprintf("http://%s:%d/internal/task", client.Host, client.Port)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.Do(req.WithContext(ctx))
	if err != nil {

		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
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

func (client *Client) SendResult(result req.Result) {
	var buf bytes.Buffer
	if result.Value == nil {
		result.Value = calculation.ErrDivisionByZero
	}
	if value, ok := result.Value.(float64); ok {
		if math.IsInf(value, 1) || math.IsInf(value, -1) || math.IsNaN(value) {
			result.Value = calculation.ErrDivisionByZero
		}
	}

	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while encoding result: %v\n", err)
		return
	}

	url := fmt.Sprintf("http://%s:%d/internal/task", client.Host, client.Port)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while creating request for posting result: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while posting result to server: %v\n", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "response status code is %d, expected %d\n", response.StatusCode, http.StatusOK)
	}
}
