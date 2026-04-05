package git

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Reader interface {
	ReadFile(ctx context.Context, repoURL, filePath string) ([]byte, error)
}

type HTTPReader struct {
	client *http.Client
}

func NewHTTPReader() *HTTPReader {
	return &HTTPReader{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

func (r *HTTPReader) ReadFile(ctx context.Context, repoURL, filePath string) ([]byte, error) {
	rawURL, err := rawFileURL(repoURL, filePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read %s: unexpected status %d", rawURL, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func rawFileURL(repoURL, filePath string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return "", fmt.Errorf("repo URL must not be empty")
	}

	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}

	cleanPath := strings.Trim(parsed.Path, "/")
	switch parsed.Host {
	case "github.com":
		parts := strings.Split(cleanPath, "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("unsupported GitHub repo URL %q", repoURL)
		}
		branch := "main"
		if len(parts) >= 4 && parts[2] == "tree" {
			branch = parts[3]
		}
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", parts[0]+"/"+parts[1], branch, path.Clean(filePath)), nil
	case "raw.githubusercontent.com":
		return fmt.Sprintf("https://%s/%s", parsed.Host, path.Join(cleanPath, filePath)), nil
	default:
		base := strings.TrimSuffix(repoURL, "/")
		if strings.Contains(base, "/-/raw/") || strings.Contains(base, "/raw/") {
			return fmt.Sprintf("%s/%s", base, path.Clean(filePath)), nil
		}
		return fmt.Sprintf("%s/raw/main/%s", base, path.Clean(filePath)), nil
	}
}
