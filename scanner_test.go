package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScanDeletesTorrentWithExcludedFiles(t *testing.T) {
	mux := http.NewServeMux()

	// Login
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ok."))
	})

	// Preferences - return one excluded pattern
	mux.HandleFunc("/api/v2/app/preferences", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preferences{
			ExcludedFileNames: "*.exe\n*.scr",
		})
	})

	// Torrent list - one torrent
	mux.HandleFunc("/api/v2/torrents/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Torrent{
			{Hash: "abc123", Name: "bad-torrent", State: "downloading"},
		})
	})

	// Torrent files - one excluded, one not
	mux.HandleFunc("/api/v2/torrents/files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TorrentFile{
			{Index: 0, Name: "setup.exe"},
			{Index: 1, Name: "readme.txt"},
		})
	})

	// Delete
	deleted := false
	mux.HandleFunc("/api/v2/torrents/delete", func(w http.ResponseWriter, r *http.Request) {
		deleted = true
		w.Write([]byte("Ok."))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Login("admin", "pass"); err != nil {
		t.Fatal(err)
	}

	cfg := Config{DeleteFiles: true}
	scan(client, cfg, false, false)

	if !deleted {
		t.Error("expected torrent to be deleted")
	}
}

func TestScanSkipsCleanTorrents(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ok."))
	})

	mux.HandleFunc("/api/v2/app/preferences", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preferences{
			ExcludedFileNames: "*.exe\n*.scr",
		})
	})

	mux.HandleFunc("/api/v2/torrents/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Torrent{
			{Hash: "abc123", Name: "clean-torrent", State: "downloading"},
		})
	})

	mux.HandleFunc("/api/v2/torrents/files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TorrentFile{
			{Index: 0, Name: "movie.mkv"},
			{Index: 1, Name: "subtitle.srt"},
		})
	})

	deleted := false
	mux.HandleFunc("/api/v2/torrents/delete", func(w http.ResponseWriter, r *http.Request) {
		deleted = true
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Login("admin", "pass"); err != nil {
		t.Fatal(err)
	}

	cfg := Config{DeleteFiles: true}
	scan(client, cfg, false, false)

	if deleted {
		t.Error("expected clean torrent to NOT be deleted")
	}
}

func TestScanDryRunDoesNotDelete(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ok."))
	})

	mux.HandleFunc("/api/v2/app/preferences", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Preferences{
			ExcludedFileNames: "*.exe",
		})
	})

	mux.HandleFunc("/api/v2/torrents/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Torrent{
			{Hash: "abc123", Name: "bad-torrent", State: "downloading"},
		})
	})

	mux.HandleFunc("/api/v2/torrents/files", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TorrentFile{
			{Index: 0, Name: "setup.exe"},
		})
	})

	deleted := false
	mux.HandleFunc("/api/v2/torrents/delete", func(w http.ResponseWriter, r *http.Request) {
		deleted = true
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL)
	if err := client.Login("admin", "pass"); err != nil {
		t.Fatal(err)
	}

	cfg := Config{DeleteFiles: true}
	scan(client, cfg, true, false) // dryRun = true

	if deleted {
		t.Error("dry run should not delete")
	}
}
