package main

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"*.scr", "setup.scr", true},
		{"*.scr", "setup.exe", false},
		{"*.exe", "malware.exe", true},
		{"foo*.exe", "foobar.exe", true},
		{"foo*.exe", "foo.exe", true},
		{"foo*.exe", "fooxbar.exe", true},
		{"foo*.exe", "foobar.txt", false},
		{"*", "anything.txt", true},
		{"", "", true},
		{"*.scr", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestMatchesAnyPattern(t *testing.T) {
	patterns := []string{"*.scr", "*.exe", "*.bat"}

	if !matchesAnyPattern("setup.scr", patterns) {
		t.Error("expected setup.scr to match")
	}
	if !matchesAnyPattern("malware.exe", patterns) {
		t.Error("expected malware.exe to match")
	}
	if matchesAnyPattern("video.mp4", patterns) {
		t.Error("expected video.mp4 to not match")
	}
	if matchesAnyPattern("readme.txt", patterns) {
		t.Error("expected readme.txt to not match")
	}
}

func TestMatchesAnyPatternCaseInsensitive(t *testing.T) {
	patterns := []string{"*.scr"}
	if !matchesAnyPattern("SETUP.SCR", patterns) {
		t.Error("expected case-insensitive match via matchesAnyPattern")
	}
}
