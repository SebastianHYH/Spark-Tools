package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Parse decodes the raw spark-tracks feed into a sorted slice of tracks.
//
// The feed is a top-level JSON object whose keys are track slugs, but it also
// carries metadata keys (e.g. "_title", "lastModified") whose values are not
// track objects. We decode into raw messages first so those entries can be
// skipped without failing the whole parse, then sort by title for stable output.
func Parse(data []byte) ([]Track, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing feed: %w", err)
	}

	tracks := make([]Track, 0, len(raw))
	for key, msg := range raw {
		// Metadata keys are underscore-prefixed; "lastModified" is the lone exception.
		if strings.HasPrefix(key, "_") || key == "lastModified" {
			continue
		}

		var entry TrackEntry
		if err := json.Unmarshal(msg, &entry); err != nil {
			// Not a track-shaped entry; skip rather than abort.
			continue
		}
		if entry.Track.Title == "" {
			continue
		}

		tracks = append(tracks, entry.Track)
	}

	sort.Slice(tracks, func(i, j int) bool {
		return strings.ToLower(tracks[i].Title) < strings.ToLower(tracks[j].Title)
	})

	return tracks, nil
}

// Find returns the first track whose title or artist contains query
// (case-insensitive), and whether a match was found.
func Find(tracks []Track, query string) (Track, bool) {
	q := strings.ToLower(query)
	for _, t := range tracks {
		if strings.Contains(strings.ToLower(t.Title), q) ||
			strings.Contains(strings.ToLower(t.Artist), q) {
			return t, true
		}
	}
	return Track{}, false
}
