package display

import (
	"image"
	"image/color"
	"math"
)

// lanczosSupport is the radius of the Lanczos-3 kernel, in source pixels.
const lanczosSupport = 3.0

// resizeImage scales src down to exactly dstW x dstH using a separable
// Lanczos-3 filter, in two passes (horizontal, then vertical).
//
// Sampling a single source pixel per output cell (nearest-neighbor) discards
// almost the entire image at album-art sizes, so every output pixel here is a
// weighted average of the source pixels around it.
func resizeImage(src image.Image, dstW, dstH int) *image.NRGBA {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()

	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	// Horizontal pass: srcW -> dstW, height unchanged.
	tmp := image.NewNRGBA(image.Rect(0, 0, dstW, srcH))
	xWeights := filterWeights(srcW, dstW)
	for y := 0; y < srcH; y++ {
		for x := 0; x < dstW; x++ {
			var r, g, bl, a float64
			for _, w := range xWeights[x] {
				pr, pg, pb, pa := nrgbaAt(src, b.Min.X+w.index, b.Min.Y+y)
				r += float64(pr) * w.weight
				g += float64(pg) * w.weight
				bl += float64(pb) * w.weight
				a += float64(pa) * w.weight
			}
			setNRGBA(tmp, x, y, r, g, bl, a)
		}
	}

	// Vertical pass: srcH -> dstH, width unchanged.
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	yWeights := filterWeights(srcH, dstH)
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			var r, g, bl, a float64
			for _, w := range yWeights[y] {
				c := tmp.NRGBAAt(x, w.index)
				r += float64(c.R) * w.weight
				g += float64(c.G) * w.weight
				bl += float64(c.B) * w.weight
				a += float64(c.A) * w.weight
			}
			setNRGBA(dst, x, y, r, g, bl, a)
		}
	}

	return dst
}

// contribution is one source pixel's share of an output pixel.
type contribution struct {
	index  int
	weight float64
}

// filterWeights precomputes, for each of the dstLen output pixels, the source
// pixels within the kernel's radius and their normalized weights.
func filterWeights(srcLen, dstLen int) [][]contribution {
	scale := float64(srcLen) / float64(dstLen)

	// When downscaling, the kernel widens to cover every source pixel that
	// collapses into one output pixel; when upscaling it stays at its natural size.
	radius := lanczosSupport
	if scale > 1 {
		radius *= scale
	}

	out := make([][]contribution, dstLen)
	for i := range out {
		// Center of output pixel i, projected into source coordinates.
		center := (float64(i)+0.5)*scale - 0.5

		start := int(math.Ceil(center - radius))
		end := int(math.Floor(center + radius))

		var sum float64
		var contribs []contribution
		for j := start; j <= end; j++ {
			x := (float64(j) - center)
			if scale > 1 {
				x /= scale
			}
			w := lanczos(x)
			if w == 0 {
				continue
			}
			// Clamp at the edges so the kernel never runs off the image.
			idx := j
			if idx < 0 {
				idx = 0
			}
			if idx >= srcLen {
				idx = srcLen - 1
			}
			sum += w
			contribs = append(contribs, contribution{index: idx, weight: w})
		}

		// Normalize so the weights total 1 and the output keeps the source's brightness.
		if sum != 0 {
			for k := range contribs {
				contribs[k].weight /= sum
			}
		}
		out[i] = contribs
	}

	return out
}

// lanczos evaluates the Lanczos-3 kernel: sinc(x) * sinc(x/3), zero past the support.
func lanczos(x float64) float64 {
	x = math.Abs(x)
	if x < 1e-9 {
		return 1
	}
	if x >= lanczosSupport {
		return 0
	}
	px := math.Pi * x
	return lanczosSupport * math.Sin(px) * math.Sin(px/lanczosSupport) / (px * px)
}

// nrgbaAt reads a pixel as non-premultiplied 8-bit RGBA.
func nrgbaAt(img image.Image, x, y int) (uint8, uint8, uint8, uint8) {
	c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	return c.R, c.G, c.B, c.A
}

// setNRGBA writes a filtered pixel, clamping to [0, 255].
// Lanczos has negative lobes, so sums can overshoot both ends of the range.
func setNRGBA(img *image.NRGBA, x, y int, r, g, b, a float64) {
	img.SetNRGBA(x, y, color.NRGBA{
		R: clamp8(r),
		G: clamp8(g),
		B: clamp8(b),
		A: clamp8(a),
	})
}

// clamp8 rounds v to the nearest byte value within [0, 255].
func clamp8(v float64) uint8 {
	v = math.Round(v)
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
