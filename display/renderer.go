package display

import (
	"fmt"
	"strings"

	"spark-cli/models"
)

// ANSI color/style codes used throughout the renderer.
const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	italic = "\x1b[3m"

	colorPurple  = "\x1b[38;2;168;85;247m" // Fortnite Festival purple
	colorCyan    = "\x1b[38;2;34;211;238m" // Accent cyan
	colorYellow  = "\x1b[38;2;250;204;21m" // Star / highlight yellow
	colorWhite   = "\x1b[38;2;255;255;255m"
	colorGray    = "\x1b[38;2;156;163;175m"
	colorGreen   = "\x1b[38;2;74;222;128m"
	colorMagenta = "\x1b[38;2;232;121;249m"

	dividerChar = "─"
)

// RenderTrack prints a full-width terminal card for a single track:
// album art ASCII on the left, song metadata on the right.
func RenderTrack(track models.Track, termWidth int) error {
	cfg := DefaultASCIIConfig()

	// Art panel takes ~40% of terminal width; metadata panel takes the rest.
	cfg.Width = termWidth * 2 / 5
	cfg.Height = 22
	cfg.UseColor = true
	cfg.UseUnicode = true

	// Fetch and render album art (fall back to placeholder on error).
	var artLines []string
	if track.AlbumArt != "" {
		var err error
		artLines, err = FetchAndRenderASCII(track.AlbumArt, cfg)
		if err != nil {
			artLines = PlaceholderArt(cfg.Width, cfg.Height)
		}
	} else {
		artLines = PlaceholderArt(cfg.Width, cfg.Height)
	}

	infoLines := buildInfoPanel(track, termWidth-cfg.Width-3, len(artLines))

	printTwoColumn(artLines, infoLines, cfg.Width)
	return nil
}

// RenderTrackList prints a compact, scrollable list of all tracks.
func RenderTrackList(tracks []models.Track, termWidth int) {
	printHeader("FORTNITE FESTIVAL — JAM TRACKS", termWidth)
	fmt.Println()

	for i, t := range tracks {
		num := fmt.Sprintf("%3d.", i+1)

		// Difficulty bar for lead guitar as a quick visual indicator.
		diffBar := difficultyBar(t.DifficultyGuitar, 7)

		fmt.Printf(
			"%s%s%s  %s%-35s%s  %s%s%s  %s%s%s  %s%s%s\n",
			colorGray, num, reset,
			colorWhite+bold, truncate(t.Title, 35), reset,
			colorGray, truncate(t.Artist, 25), reset,
			colorYellow, diffBar, reset,
			colorGray, fmt.Sprintf("♩%d", t.BPM), reset,
		)
	}

	fmt.Println()
	printDivider(termWidth)
}

// buildInfoPanel assembles the right-hand metadata column as a slice of strings.
// targetHeight ensures the column matches the art panel height (padded if short).
func buildInfoPanel(track models.Track, width, targetHeight int) []string {
	var lines []string

	add := func(line string) {
		lines = append(lines, line)
	}

	// — Title block —
	add(colorPurple + bold + wrapText(track.Title, width) + reset)
	add(colorCyan + italic + truncate(track.Artist, width) + reset)

	if track.Album != "" {
		add(colorGray + dim + truncate(track.Album, width) + reset)
	}

	add("")
	add(colorGray + strings.Repeat(dividerChar, min(width, 30)) + reset)
	add("")

	// — Quick stats row —
	year := ""
	if track.Year > 0 {
		year = fmt.Sprintf("  %s%d%s", colorGray, track.Year, reset)
	}
	add(fmt.Sprintf(
		"%s♩ %d BPM%s%s   %s%s %s%s",
		colorYellow, track.BPM, reset,
		year,
		colorCyan, track.Key, track.Mode, reset,
	))

	add("")

	// — Difficulty table —
	add(colorWhite + bold + "Difficulty" + reset)
	add("")
	add(instrumentRow("Vocals  ", track.DifficultyVocals))
	add(instrumentRow("Guitar  ", track.DifficultyGuitar))
	add(instrumentRow("Bass    ", track.DifficultyBass))
	add(instrumentRow("Drums   ", track.DifficultyDrums))
	add(instrumentRow("Pro Lead", track.DifficultyProLead))
	add(instrumentRow("Pro Bass", track.DifficultyProBass))

	add("")
	add(colorGray + strings.Repeat(dividerChar, min(width, 30)) + reset)
	add("")

	// — Genre tags —
	if len(track.Genre) > 0 {
		tags := ""
		for _, g := range track.Genre {
			tags += colorMagenta + "  #" + g + reset
		}
		add(strings.TrimSpace(tags))
		add("")
	}

	// Pad to target height so columns align.
	for len(lines) < targetHeight {
		lines = append(lines, "")
	}

	return lines
}

// printTwoColumn zips art lines (left) and info lines (right) and prints them
// side-by-side with a gutter between.
func printTwoColumn(left, right []string, leftWidth int) {
	height := max(len(left), len(right))
	gutter := "   "

	for i := 0; i < height; i++ {
		leftLine := ""
		if i < len(left) {
			leftLine = left[i]
		}

		rightLine := ""
		if i < len(right) {
			rightLine = right[i]
		}

		// Pad left column to consistent visual width (ANSI codes don't count).
		fmt.Printf("%s%s%s\n", padVisual(leftLine, leftWidth), gutter, rightLine)
	}
}

// printHeader renders a centered, styled banner.
func printHeader(title string, width int) {
	divider := colorPurple + strings.Repeat("═", width) + reset
	pad := (width - len(title)) / 2
	centered := strings.Repeat(" ", pad) + colorYellow + bold + title + reset

	fmt.Println(divider)
	fmt.Println(centered)
	fmt.Println(divider)
}

// printDivider prints a full-width separator line.
func printDivider(width int) {
	fmt.Println(colorGray + strings.Repeat(dividerChar, width) + reset)
}

// instrumentRow renders one line of the difficulty table, e.g.:
// "Guitar   ★★★★☆☆☆  4"
func instrumentRow(label string, level int) string {
	const maxLevel = 7
	stars := ""
	for i := 1; i <= maxLevel; i++ {
		if i <= level {
			stars += colorYellow + "★" + reset
		} else {
			stars += colorGray + "☆" + reset
		}
	}
	levelStr := fmt.Sprintf("%s%d%s", colorGray, level, reset)
	return fmt.Sprintf("  %s%s%s  %s  %s", colorGray, label, reset, stars, levelStr)
}

// difficultyBar returns a compact bar like "████░░░" for use in list view.
func difficultyBar(level, max int) string {
	filled := strings.Repeat("█", level)
	empty := strings.Repeat("░", max-level)
	return filled + empty
}

// truncate shortens s to maxLen runes, adding "…" if cut.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// wrapText naively wraps s at width characters (rune-aware).
func wrapText(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}

// padVisual pads a string to targetWidth visible characters.
// Because ANSI escape codes are invisible, we strip them when counting width.
func padVisual(s string, targetWidth int) string {
	visible := visibleLen(s)
	if visible >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-visible)
}

// visibleLen counts the number of printable runes in s, ignoring ANSI escapes.
func visibleLen(s string) int {
	count := 0
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			count++
		}
	}
	return count
}

// min / max helpers (pre-Go 1.21 safe).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
