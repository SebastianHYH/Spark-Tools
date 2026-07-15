package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const baseURL = "https://fortnitecontent-website-prod07.ol.epicgames.com/content/api/pages/fortnite-game/spark-tracks"
const localPath = "./data/spark-tracks.json"

/*
Fetch retrieves the spark-tracks feed.

It tries the live Epic endpoint first and refreshes the local cache on success.
If the remote request fails it falls back to the cached copy on disk, so the
app keeps working offline. The raw JSON body is returned to the caller.
*/
func Fetch() ([]byte, error) {
	body, err := fetchFromRemote()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Remote fetch failed:", err)
		fmt.Fprintln(os.Stderr, "Falling back to local cache...")
		return fetchFromLocal()
	}

	if err := saveToLocal(body); err != nil {
		// A stale-but-usable cache is better than failing the whole run.
		fmt.Fprintln(os.Stderr, "Warning: could not update local cache:", err)
	}

	return body, nil
}

/*
fetchFromRemote performs the HTTP GET against baseURL.
*/
func fetchFromRemote() ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

/*
saveToLocal writes the feed to the local cache file, creating the directory if needed.
*/
func saveToLocal(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(localPath, data, 0o644)
}

/*
fetchFromLocal reads the feed from the local cache file.
*/
func fetchFromLocal() ([]byte, error) {
	body, err := os.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("no remote connection and no local cache available: %w", err)
	}
	return body, nil
}
