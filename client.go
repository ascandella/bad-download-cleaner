package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Client struct {
	http *http.Client
	url  string
}

type Torrent struct {
	Hash  string `json:"hash"`
	Name  string `json:"name"`
	State string `json:"state"`
	Size  int64  `json:"size"`
}

type TorrentFile struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Progress float64 `json:"progress"`
	Priority int    `json:"priority"`
}

type Preferences struct {
	ExcludedFileNames string `json:"excluded_file_names"`
}

func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{Jar: jar},
		url:  strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) Login(username, password string) error {
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	req, err := http.NewRequest("POST", c.url+"/api/v2/auth/login", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.url)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(body) != "Ok." {
		return fmt.Errorf("login failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) GetExcludedFileNames() ([]string, error) {
	resp, err := c.http.Get(c.url + "/api/v2/app/preferences")
	if err != nil {
		return nil, fmt.Errorf("getting preferences: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get preferences failed (status %d): %s", resp.StatusCode, string(body))
	}

	var prefs Preferences
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return nil, fmt.Errorf("decoding preferences: %w", err)
	}

	if prefs.ExcludedFileNames == "" {
		return nil, nil
	}

	var patterns []string
	for _, line := range strings.Split(prefs.ExcludedFileNames, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			patterns = append(patterns, line)
		}
	}
	return patterns, nil
}

func (c *Client) GetTorrents() ([]Torrent, error) {
	resp, err := c.http.Get(c.url + "/api/v2/torrents/info")
	if err != nil {
		return nil, fmt.Errorf("getting torrents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get torrents failed (status %d): %s", resp.StatusCode, string(body))
	}

	var torrents []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&torrents); err != nil {
		return nil, fmt.Errorf("decoding torrents: %w", err)
	}
	return torrents, nil
}

func (c *Client) GetTorrentFiles(hash string) ([]TorrentFile, error) {
	resp, err := c.http.Get(c.url + "/api/v2/torrents/files?hash=" + hash)
	if err != nil {
		return nil, fmt.Errorf("getting files for %s: %w", hash, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get files failed (status %d): %s", resp.StatusCode, string(body))
	}

	var files []TorrentFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("decoding files: %w", err)
	}
	return files, nil
}

func (c *Client) DeleteTorrent(hash string, deleteFiles bool) error {
	data := url.Values{
		"hashes":      {hash},
		"deleteFiles": {fmt.Sprintf("%t", deleteFiles)},
	}

	req, err := http.NewRequest("POST", c.url+"/api/v2/torrents/delete", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}
