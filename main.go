package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"spark-cli/api"
	"spark-cli/display"
	"spark-cli/models"
)

func main() {
	data, err := api.Fetch()
	if err != nil {
		log.Fatal(err)
	}

	tracks, err := models.Parse(data)
	if err != nil {
		log.Fatal(err)
	}
	if len(tracks) == 0 {
		log.Fatal("no tracks found in feed")
	}

	width := terminalWidth()

	// With a query argument, show the matching track's full card.
	// Without one, show the compact list of every track.
	if len(os.Args) > 1 {
		query := strings.Join(os.Args[1:], " ")
		track, ok := models.Find(tracks, query)
		if !ok {
			log.Fatalf("no track matching %q", query)
		}
		if err := display.RenderTrack(track, width); err != nil {
			log.Fatal(err)
		}
		return
	}

	display.RenderTrackList(tracks, width)
}

// terminalWidth reports the usable column count, falling back to a readable
// default when it can't be determined (e.g. output piped to a file).
func terminalWidth() int {
	const fallback = 100
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(cols)); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
