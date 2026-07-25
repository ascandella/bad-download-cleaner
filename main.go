package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
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
		scan(client, cfg, *dryRun, *verbose)
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
			scan(client, cfg, *dryRun, *verbose)
		case s := <-sig:
			log.Printf("Received %s, shutting down", s)
			return
		}
	}
}
