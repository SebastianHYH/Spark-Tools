package display

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
)

// ASCIIConfig controls how the album art is rendered in the terminal.
type ASCIIConfig struct {
	Width    int  // Character width of the output block
	Height   int  // Character height; 0 derives it from the image's aspect ratio
	UseColor bool // Whether to emit ANSI 24-bit color escape codes
	Complex  bool // Use the detailed 69-char ramp instead of the simple 10-char one
}

// DefaultASCIIConfig returns sensible defaults for a side-panel album art block.
func DefaultASCIIConfig() ASCIIConfig {
	return ASCIIConfig{
		Width:    40,
		Height:   0,
		UseColor: true,
		Complex:  true,
	}
}

// Character ramps, ordered darkest pixel to brightest.
// Taken from ascii-image-converter, which sources them from
// http://paulbourke.net/dataformats/asciiart/
const (
	rampSimple   = " .:-=+*#%@"
	rampDetailed = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
)

// cellAspect corrects for terminal characters being roughly twice as tall
// as they are wide, so square images don't come out stretched.
const cellAspect = 0.5

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
//
// The image is first resized down to exactly one pixel per character cell, then
// each pixel becomes a ramp character chosen by its brightness, tinted with the
// pixel's own color.
func renderImage(img image.Image, cfg ASCIIConfig) []string {
	w, h := artDimensions(img, cfg)
	small := resizeImage(img, w, h)

	ramp := rampSimple
	if cfg.Complex {
		ramp = rampDetailed
	}

	lines := make([]string, 0, h)
	for y := 0; y < h; y++ {
		var sb strings.Builder

		for x := 0; x < w; x++ {
			r, g, b := rgbAt(small, x, y)
			ch := rampChar(ramp, grayLevel(r, g, b))

			if cfg.UseColor {
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm%c", r, g, b, ch)
			} else {
				sb.WriteRune(ch)
			}
		}

		if cfg.UseColor {
			sb.WriteString(reset)
		}
		lines = append(lines, sb.String())
	}

	return lines
}

// artDimensions returns the character grid to rasterize into. An explicit
// cfg.Height wins; otherwise the height follows the image's aspect ratio,
// corrected for the shape of a terminal cell.
func artDimensions(img image.Image, cfg ASCIIConfig) (int, int) {
	w := cfg.Width
	if w < 1 {
		w = 1
	}

	h := cfg.Height
	if h < 1 {
		b := img.Bounds()
		aspect := float64(b.Dx()) / float64(b.Dy())
		h = int(cellAspect * float64(w) / aspect)
		if h < 1 {
			h = 1
		}
	}

	return w, h
}

// rgbAt returns the (r, g, b) bytes of the pixel at (x, y), handling
// the image's own bounds offset.
func rgbAt(img image.Image, x, y int) (uint8, uint8, uint8) {
	b := img.Bounds()
	c := img.At(b.Min.X+x, b.Min.Y+y)
	r, g, bv, _ := color.NRGBAModel.Convert(c).RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(bv >> 8)
}

// grayLevel converts RGB to a perceptual brightness in [0, 255],
// using the BT.601 luma coefficients behind color.GrayModel.
func grayLevel(r, g, b uint8) uint8 {
	return color.GrayModel.Convert(color.NRGBA{R: r, G: g, B: b, A: 255}).(color.Gray).Y
}

// rampChar maps a brightness in [0, 255] to a character in the ramp.
func rampChar(ramp string, level uint8) rune {
	// Pure white would land one past the end of the ramp.
	if level == 255 {
		return rune(ramp[len(ramp)-1])
	}
	idx := int(float64(level) / 255.0 * float64(len(ramp)))
	return rune(ramp[idx])
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
