package models

type Track struct {
	Title    string `json:"tt"`
	Artist   string `json:"an"`
	Album    string `json:"ab"`
	Year     int    `json:"ry"`
	AlbumArt string `json:"au"`
}

/*
 * Go structs that map to the API response
 */
