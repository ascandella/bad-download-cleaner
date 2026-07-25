package main

import (
	"log"
	"path/filepath"
)

func scan(client *Client, cfg Config, dryRun bool, verbose bool) {
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

	log.Println("Done checking torrents")
}
