# bad-download-cleaner

A daemon that monitors qBittorrent for downloads that will never progress because all their files match the server's excluded file name patterns, then deletes them.

## How it works

1. Queries qBittorrent for its `excluded_file_names` setting
2. For each active torrent, checks if **every** file matches an excluded pattern
3. If so, deletes the torrent (and its files from disk)

## Usage

```bash
# Run once and exit
./bad-download-cleaner --once

# Daemon mode, poll every 60 seconds
./bad-download-cleaner --interval 60

# Preview what would be deleted
./bad-download-cleaner --once --dry-run
```

## Configuration

| Env var | Default | Description |
|---|---|---|
| `QB_URL` | `http://localhost:8080` | qBittorrent WebUI URL |
| `QB_USER` | `admin` | WebUI username |
| `QB_PASS` | `adminadmin` | WebUI password |
| `QB_DELETE_FILES` | `true` | Delete files from disk (not just the torrent) |

Copy `.env.example` to `.env` for local development.

## Docker

```bash
# Build
docker build -t bad-download-cleaner .

# Run
docker run --rm bad-download-cleaner --once
```

Or pull from GHCR:

```bash
docker run --rm \
  -e QB_URL=http://qbittorrent:8080 \
  -e QB_USER=admin \
  -e QB_PASS=secret \
  ghcr.io/ascandella/bad-download-cleaner:latest --once
```

## How to configure excluded files in qBittorrent

In qBittorrent go to **Tools > Options > Downloads > Exclude file names** and add patterns like:

```
*.scr
*.exe
*.bat
```

This tool reads that exact list and uses it to identify stuck downloads.
