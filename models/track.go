package models

type SparkTracksResponse map[string]TrackEntry

// TrackEntry wraps the track data returned per key in the response.
type TrackEntry struct {
	Track    Track  `json:"track"`
	ID       string `json:"id"`
	DevName  string `json:"_devName"`
	NoIndex  bool   `json:"_noIndex"`
	Locale   string `json:"_locale"`
}

// Track holds all metadata for a single Fortnite Festival jam track.
type Track struct {
	// Core identity
	ID       string `json:"id"`        // Internal track ID

	// Song info
	Title    string `json:"tt"`        // Track title
	Artist   string `json:"an"`        // Artist name
	Album    string `json:"ab"`        // Album name
	Year     int    `json:"ry"`        // Release year
	Duration int    `json:"dn"`        // Duration in seconds (note: may conflict with DevName key, verify on live data)

	// Audio/music properties
	BPM      int    `json:"bpm"`       // Beats per minute
	Key      string `json:"mk"`        // Musical key (e.g. "C", "F#")
	Mode     string `json:"mu"`        // Scale mode (e.g. "major", "minor")

	// Art / media
	AlbumArt string `json:"au"`        // URL to album art image

	// Availability
	ReleaseDate  string `json:"jr"`    // Date added to Fortnite Festival
	LastModified string `json:"jrm"`   // Last time the entry was modified

	// Difficulty ratings per instrument (1–7 scale)
	DifficultyVocals  int `json:"dv"`  // Vocals difficulty
	DifficultyGuitar  int `json:"dd"`  // Lead guitar difficulty
	DifficultyBass    int `json:"db"`  // Bass difficulty
	DifficultyDrums   int `json:"dp"`  // Drums difficulty
	DifficultyProLead int `json:"dl"`  // Pro lead difficulty
	DifficultyProBass int `json:"da"`  // Pro bass difficulty

	// Genre / tags
	Genre []string `json:"ge"`         // Genre tags (e.g. ["Pop", "Rock"])
}

