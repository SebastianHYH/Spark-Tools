package display

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"strings"
)

// ASCIIConfig controls how the album art is rendered in the terminal.
type ASCIIConfig struct {
	Width      int  // Character width of the output block
	Height     int  // Character height of the output block
	UseColor   bool // Whether to emit ANSI 24-bit color escape codes
	UseUnicode bool // Use block characters (▄) instead of ASCII density chars
}

// DefaultASCIIConfig returns sensible defaults for a side-panel album art block.
func DefaultASCIIConfig() ASCIIConfig {
	return ASCIIConfig{
		Width:      40,
		Height:     20,
		UseColor:   true,
		UseUnicode: true,
	}
}

// density is the ASCII character ramp from darkest to brightest.
// Used when UseUnicode is false.
const density = " .,:;i1tfLCG08@"

// FetchAndRenderASCII downloads an image from imageURL and converts it to
// an ASCII/ANSI art string ready to print in the terminal.
func FetchAndRenderASCII(imageURL string, cfg ASCIIConfig) ([]string, error) {
	img, err := fetchImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	return renderImage(img, cfg), nil
}

// fetchImage downloads and decodes an image from a URL.
func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return decodeImage(resp.Body)
}

// decodeImage decodes an image.Image from a reader.
// Supports JPEG and PNG via the blank imports at the top.
func decodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return img, nil
}

// renderImage converts an image.Image into a slice of terminal-printable strings,
// one string per row. The caller can print them line by line.
func renderImage(img image.Image, cfg ASCIIConfig) []string {
	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	// Each terminal character is roughly 2× taller than it is wide,
	// so we halve the target height to preserve aspect ratio.
	targetH := cfg.Height
	targetW := cfg.Width

	lines := make([]string, 0, targetH)

	if cfg.UseUnicode && cfg.UseColor {
		// Half-block mode: each character row covers 2 pixel rows (top ▄ bottom).
		// We sample pairs of rows and encode them as fg (lower) + bg (upper) color.
		lines = renderHalfBlock(img, srcW, srcH, targetW, targetH)
	} else if cfg.UseColor {
		lines = renderColorASCII(img, srcW, srcH, targetW, targetH)
	} else {
		lines = renderMonoASCII(img, srcW, srcH, targetW, targetH)
	}

	return lines
}

// renderHalfBlock uses the Unicode LOWER HALF BLOCK (▄) character.
// The foreground color is the lower pixel, the background is the upper pixel.
// This effectively doubles vertical resolution compared to plain ASCII.
func renderHalfBlock(img image.Image, srcW, srcH, targetW, targetH int) []string {
	lines := make([]string, 0, targetH)

	for row := 0; row < targetH; row++ {
		var sb strings.Builder

		// Two source pixel rows map to one output character row.
		upperRowY := int(math.Floor(float64(row*2) * float64(srcH) / float64(targetH*2)))
		lowerRowY := int(math.Floor(float64(row*2+1) * float64(srcH) / float64(targetH*2)))

		for col := 0; col < targetW; col++ {
			srcX := int(math.Floor(float64(col) * float64(srcW) / float64(targetW)))

			upperR, upperG, upperB := rgbAt(img, srcX, upperRowY)
			lowerR, lowerG, lowerB := rgbAt(img, srcX, lowerRowY)

			// BG = upper pixel, FG = lower pixel, character = ▄
			sb.WriteString(fmt.Sprintf(
				"\x1b[48;2;%d;%d;%dm\x1b[38;2;%d;%d;%dm▄",
				upperR, upperG, upperB,
				lowerR, lowerG, lowerB,
			))
		}

		// Reset color at end of each line.
		sb.WriteString("\x1b[0m")
		lines = append(lines, sb.String())
	}

	return lines
}

// renderColorASCII maps brightness to a density character and colors it
// with the pixel's actual ANSI foreground color.
func renderColorASCII(img image.Image, srcW, srcH, targetW, targetH int) []string {
	lines := make([]string, 0, targetH)

	for row := 0; row < targetH; row++ {
		var sb strings.Builder
		srcY := int(math.Floor(float64(row) * float64(srcH) / float64(targetH)))

		for col := 0; col < targetW; col++ {
			srcX := int(math.Floor(float64(col) * float64(srcW) / float64(targetW)))
			r, g, b := rgbAt(img, srcX, srcY)
			brightness := luminance(r, g, b)
			ch := densityChar(brightness)
			sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", r, g, b, ch))
		}

		sb.WriteString("\x1b[0m")
		lines = append(lines, sb.String())
	}

	return lines
}

// renderMonoASCII maps brightness to a density character, no color codes.
func renderMonoASCII(img image.Image, srcW, srcH, targetW, targetH int) []string {
	lines := make([]string, 0, targetH)

	for row := 0; row < targetH; row++ {
		var sb strings.Builder
		srcY := int(math.Floor(float64(row) * float64(srcH) / float64(targetH)))

		for col := 0; col < targetW; col++ {
			srcX := int(math.Floor(float64(col) * float64(srcW) / float64(targetW)))
			r, g, b := rgbAt(img, srcX, srcY)
			brightness := luminance(r, g, b)
			sb.WriteRune(densityChar(brightness))
		}

		lines = append(lines, sb.String())
	}

	return lines
}

// rgbAt returns the (r, g, b) bytes of the pixel at (x, y), handling
// the image's own bounds offset.
func rgbAt(img image.Image, x, y int) (uint8, uint8, uint8) {
	b := img.Bounds()
	c := img.At(b.Min.X+x, b.Min.Y+y)
	r, g, bv, _ := color.NRGBAModel.Convert(c).RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(bv >> 8)
}

// luminance converts RGB to a perceptual brightness value in [0, 1].
// Uses the BT.601 luma coefficients.
func luminance(r, g, b uint8) float64 {
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
}

// densityChar maps a brightness in [0, 1] to a character in the density ramp.
func densityChar(brightness float64) rune {
	idx := int(brightness * float64(len(density)-1))
	if idx >= len(density) {
		idx = len(density) - 1
	}
	return rune(density[idx])
}

// PlaceholderArt returns a simple bordered placeholder block when
// no image URL is available or the fetch fails.
func PlaceholderArt(width, height int) []string {
	lines := make([]string, height)
	border := strings.Repeat("─", width-2)

	lines[0] = "┌" + border + "┐"
	lines[height-1] = "└" + border + "┘"

	mid := height / 2
	for i := 1; i < height-1; i++ {
		if i == mid {
			label := " NO IMAGE "
			pad := (width - 2 - len(label)) / 2
			left := strings.Repeat(" ", pad)
			right := strings.Repeat(" ", width-2-pad-len(label))
			lines[i] = "│" + left + label + right + "│"
		} else {
			lines[i] = "│" + strings.Repeat(" ", width-2) + "│"
		}
	}

	return lines
}
