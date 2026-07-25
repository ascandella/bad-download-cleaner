package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	once := flag.Bool("once", false, "Run once and exit (default: poll continuously)")
	pollSec := flag.Int("interval", 30, "Poll interval in seconds (daemon mode only)")
	dryRun := flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
	verbose := flag.Bool("verbose", false, "Log per-file match details")
	flag.Parse()

	_ = godotenv.Load()

	cfg := LoadConfig()
	if *pollSec > 0 {
		cfg.PollIntervalSec = *pollSec
	}

	log.Printf("Delete files from disk: %v", cfg.DeleteFiles)
	log.Printf("qBittorrent URL: %s", cfg.URL)

	client := NewClient(cfg.URL)
	if err := client.Login(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	log.Println("Logged in successfully")

	if *once {
		run(client, cfg, *dryRun, *verbose)
		return
	}

	log.Printf("Polling every %d seconds...", cfg.PollIntervalSec)
	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSec) * time.Second)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			run(client, cfg, *dryRun, *verbose)
		case s := <-sig:
			log.Printf("Received %s, shutting down", s)
			return
		}
	}
}

func run(client *Client, cfg Config, dryRun bool, verbose bool) {
	patterns, err := client.GetExcludedFileNames()
	if err != nil {
		log.Printf("Error fetching excluded file names: %v", err)
		return
	}

	if len(patterns) == 0 {
		log.Println("No excluded file names configured in qBittorrent - nothing to check")
		return
	}

	log.Printf("qBittorrent excluded file name patterns: %v", patterns)

	torrents, err := client.GetTorrents()
	if err != nil {
		log.Printf("Error fetching torrents: %v", err)
		return
	}

	if len(torrents) == 0 {
		log.Println("No active torrents")
		return
	}

	log.Printf("Checking %d torrent(s)...", len(torrents))

	for _, t := range torrents {
		files, err := client.GetTorrentFiles(t.Hash)
		if err != nil {
			log.Printf("Error fetching files for %s (%s): %v", t.Name, t.Hash[:8], err)
			continue
		}

		if len(files) == 0 {
			continue
		}

		var excludedFiles []TorrentFile
		for _, f := range files {
			base := filepath.Base(f.Name)
			matched := matchesAnyPattern(base, patterns)
			if verbose {
				log.Printf("  %s (base: %s) excluded=%v", f.Name, base, matched)
			}
			if matched {
				excludedFiles = append(excludedFiles, f)
			}
		}

		if len(excludedFiles) == 0 {
			continue
		}

		log.Printf("EXCLUDED FILES FOUND in %s:", t.Name)
		for _, f := range excludedFiles {
			log.Printf("  - %s", f.Name)
		}

		if dryRun {
			log.Printf("[DRY RUN] Would delete torrent: %s", t.Name)
			continue
		}

		if err := client.DeleteTorrent(t.Hash, cfg.DeleteFiles); err != nil {
			log.Printf("Error deleting %s: %v", t.Name, err)
			continue
		}
		log.Printf("Deleted torrent: %s", t.Name)
	}
}

func matchesAnyPattern(filename string, patterns []string) bool {
	lower := strings.ToLower(filename)
	for _, p := range patterns {
		if matchGlob(strings.ToLower(p), lower) {
			return true
		}
	}
	return false
}

func matchGlob(pattern, name string) bool {
	for {
		if pattern == "" {
			return name == ""
		}
		if pattern == "*" {
			return true
		}

		i := strings.Index(pattern, "*")
		if i < 0 {
			return strings.HasSuffix(name, pattern)
		}

		prefix := pattern[:i]
		if !strings.HasPrefix(name, prefix) {
			return false
		}

		name = name[len(prefix):]
		pattern = pattern[i+1:]
	}
}
