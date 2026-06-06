package restclient

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: timeout,
        },
    }
}

func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
    url := c.baseURL + path
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
    url := c.baseURL + path
    
    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
    }
    
    if result != nil {
        return json.NewDecoder(resp.Body).Decode(result)
    }
    return nil
}

func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
    url := c.baseURL + path
    
    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return err
    }
    
    req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    if result != nil {
        return json.NewDecoder(resp.Body).Decode(result)
    }
    return nil
}

func (c *Client) Delete(ctx context.Context, path string) error {
    url := c.baseURL + path
    
    req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
    if err != nil {
        return err
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    return nil
}