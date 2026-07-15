package models

type SparkTracksResponse map[string]TrackEntry

// TrackEntry wraps the track data returned per key in the response.
type TrackEntry struct {
	Track   Track  `json:"track"`
	ID      string `json:"id"`
	DevName string `json:"_devName"`
	NoIndex bool   `json:"_noIndex"`
	Locale  string `json:"_locale"`
}

// Track holds all metadata for a single Fortnite Festival jam track.
//
// The live feed uses terse two/three-letter field names. These were verified
// against data/spark-tracks.json — they are not guessable, so confirm against
// real data before adding more.
type Track struct {
	// Song info
	Title    string `json:"tt"` // Track title
	Artist   string `json:"an"` // Artist name
	Album    string `json:"ab"` // Album name
	Year     int    `json:"ry"` // Release year
	Duration int    `json:"dn"` // Duration in seconds

	// Audio/music properties
	BPM  int    `json:"mt"` // Tempo (beats per minute)
	Key  string `json:"mk"` // Musical key (e.g. "C", "F#")
	Mode string `json:"mm"` // Scale mode (e.g. "Major", "Minor")

	// Art / media
	AlbumArt string `json:"au"` // URL to album art image

	// Per-instrument difficulty ratings, nested under the feed's "in" object.
	Intensities Intensities `json:"in"`

	// Genre / tags
	Genre []string `json:"ge"` // Genre tags (e.g. ["Pop", "Rock"])
}

// Intensities holds the per-instrument difficulty ratings for a track.
// Values run 0–6 in the feed (commonly shown as a 1–7 star scale).
type Intensities struct {
	Vocals   int `json:"vl"` // Vocals
	Guitar   int `json:"gr"` // Lead guitar
	Bass     int `json:"ba"` // Bass
	Drums    int `json:"ds"` // Drums
	ProLead  int `json:"pg"` // Pro lead (pro guitar)
	ProBass  int `json:"pb"` // Pro bass
	ProDrums int `json:"pd"` // Pro drums
	Band     int `json:"bd"` // Overall band intensity
}
