// Package modelrepo provides model repository integration for HuggingFace and ModelScope
package modelrepo

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"encoding/json"
	"strings"
)

// Source represents the model repository source
type Source string

const (
	SourceHuggingFace Source = "huggingface"
	SourceModelScope  Source = "modelscope"
)

// Client is a model repository client
type Client struct {
	httpClient *http.Client
	hfToken    string // Optional HuggingFace authentication token
}

// NewClient creates a new model repository client
func NewClient() *Client {
	// 设置合理的超时时间
	timeout := 10 * time.Second // 10 秒超时，避免用户等待太久

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second, // 连接超时
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// SetHFToken sets the HuggingFace authentication token
func (c *Client) SetHFToken(token string) {
	c.hfToken = token
}

// FileInfo represents information about a model file
type FileInfo struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
}

// GenerateDownloadURL generates a download URL from repository information
func (c *Client) GenerateDownloadURL(source Source, repoID, fileName string) (string, error) {
	switch source {
	case SourceHuggingFace:
		return c.generateHuggingFaceURL(repoID, fileName)
	case SourceModelScope:
		return c.generateModelScopeURL(repoID, fileName)
	default:
		return "", fmt.Errorf("unsupported source: %s", source)
	}
}

// generateHuggingFaceURL generates a HuggingFace download URL
func (c *Client) generateHuggingFaceURL(repoID, fileName string) (string, error) {
	// HuggingFace URL format: https://huggingface.co/{repoId}/resolve/main/{fileName}
	base := fmt.Sprintf("https://huggingface.co/%s/resolve/main/", repoID)
	if fileName != "" {
		base += fileName
	}
	return base, nil
}

// generateModelScopeURL generates a ModelScope download URL
func (c *Client) generateModelScopeURL(repoID, fileName string) (string, error) {
	// ModelScope URL format: https://www.modelscope.cn/api/v1/models/{repoId}/repo?Revision=master&FilePath={fileName}
	// For direct download: https://www.modelscope.cn/models/{repoId}/master/{fileName}
	base := fmt.Sprintf("https://www.modelscope.cn/api/v1/models/%s/repo?Revision=master&FilePath=", repoID)
	if fileName != "" {
		base += fileName
	}
	return base, nil
}

// ListGGUFFiles lists GGUF files in a HuggingFace repository
func (c *Client) ListGGUFFiles(repoID string) ([]FileInfo, error) {
	url := fmt.Sprintf("https://huggingface.co/api/models/%s?tree=1&recursive=1", repoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add auth token if available
	if c.hfToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.hfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch model info: %s", resp.Status)
	}

	var result struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		} `json:"tree"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, item := range result.Tree {
		// Only include files (not directories)
		if item.Type == "file" && isGGUFFile(item.Path) {
			files = append(files, FileInfo{
				Name:        item.Path,
				Size:        item.Size,
				DownloadURL: fmt.Sprintf("https://huggingface.co/%s/resolve/main/%s", repoID, item.Path),
			})
		}
	}

	return files, nil
}

// isGGUFFile checks if a file is a GGUF model file
func isGGUFFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".gguf")
}

// ParseRepoID validates and parses a repository ID
func ParseRepoID(repoID string) (owner, model string, err error) {
	parts := strings.Split(repoID, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo ID format (expected 'owner/model'): %s", repoID)
	}
	return parts[0], parts[1], nil
}
