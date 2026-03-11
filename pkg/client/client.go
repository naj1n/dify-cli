package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(host, apiKey string) *Client {
	base := strings.TrimRight(host, "/")
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	return &Client{
		baseURL:    base,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	u := c.baseURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil && method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

// --- Workflow API ---

type RunWorkflowRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
	Files        []interface{}          `json:"files,omitempty"`
}

func (c *Client) RunWorkflow(inputs map[string]interface{}, user, responseMode string) ([]byte, error) {
	body := RunWorkflowRequest{
		Inputs:       inputs,
		ResponseMode: responseMode,
		User:         user,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(http.MethodPost, "/workflows/run", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

type SSEEvent struct {
	Event string
	Data  string
}

type StreamCallback func(event SSEEvent)

func (c *Client) RunWorkflowStream(inputs map[string]interface{}, user string, callback StreamCallback) error {
	body := RunWorkflowRequest{
		Inputs:       inputs,
		ResponseMode: "streaming",
		User:         user,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := c.newRequest(http.MethodPost, "/workflows/run", bytes.NewReader(data))
	if err != nil {
		return err
	}

	streamClient := &http.Client{Timeout: 0}
	resp, err := streamClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(errBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		jsonData := strings.TrimPrefix(line, "data: ")
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
			continue
		}
		eventName, _ := parsed["event"].(string)
		callback(SSEEvent{Event: eventName, Data: jsonData})
	}
	return scanner.Err()
}

func (c *Client) GetWorkflowRunDetail(workflowRunID string) ([]byte, error) {
	req, err := c.newRequest(http.MethodGet, "/workflows/run/"+workflowRunID, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) StopWorkflow(taskID, user string) ([]byte, error) {
	body := map[string]string{"user": user}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(http.MethodPost, "/workflows/tasks/"+taskID+"/stop", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) GetWorkflowLogs(params map[string]string) ([]byte, error) {
	q := url.Values{}
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	path := "/workflows/logs"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) UploadFile(filePath, user string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := writer.WriteField("user", user); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	u := c.baseURL + "/files/upload"
	req, err := http.NewRequest(http.MethodPost, u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return c.do(req)
}

func (c *Client) GetAppInfo() ([]byte, error) {
	req, err := c.newRequest(http.MethodGet, "/info", nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) GetParameters() ([]byte, error) {
	req, err := c.newRequest(http.MethodGet, "/parameters", nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) GetSite() ([]byte, error) {
	req, err := c.newRequest(http.MethodGet, "/site", nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}
