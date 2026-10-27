// genicon draws the app icon (Mars with a tilted ring on a dark rounded square) with the standard
// library only and writes build/appicon.png plus build/windows/icon.ico (PNG-compressed entries).
//
//	go run ./tools/genicon
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	size = 1024
	ss   = 3 // supersampling per axis
)

type rgb struct{ r, g, b float64 }

func mix(a, b rgb, t float64) rgb {
	t = math.Max(0, math.Min(1, t))
	return rgb{a.r + (b.r-a.r)*t, a.g + (b.g-a.g)*t, a.b + (b.b-a.b)*t}
}

func over(dst, src rgb, a float64) rgb { return mix(dst, src, a) }

func smooth(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

// roundedRectCoverage is 1 inside a rounded square covering the canvas, 0 outside.
func inRoundedRect(x, y float64) bool {
	const m, r = 0.0, 0.225
	dx := math.Max(math.Abs(x-0.5)-(0.5-m-r), 0)
	dy := math.Max(math.Abs(y-0.5)-(0.5-m-r), 0)
	return dx*dx+dy*dy <= r*r
}

func shade(x, y float64) (rgb, float64) { // x,y in 0..1; returns colour and alpha
	if !inRoundedRect(x, y) {
		return rgb{}, 0
	}
	// background: near-black with a red glow behind the planet
	bg := rgb{0.06, 0.06, 0.06}
	d := math.Hypot(x-0.48, y-0.52)
	bg = mix(bg, rgb{0.45, 0.02, 0.02}, 0.55*math.Exp(-d*d*9))

	// tilted ring frame
	const tilt = -18 * math.Pi / 180
	px, py := x-0.5, y-0.5
	rx := px*math.Cos(-tilt) - py*math.Sin(-tilt)
	ry := px*math.Sin(-tilt) + py*math.Cos(-tilt)
	const ringA, ringB = 0.43, 0.125
	ringF := math.Sqrt((rx/ringA)*(rx/ringA) + (ry/ringB)*(ry/ringB))
	ringW := 0.010
	ringCover := 1 - smooth(ringW*0.5, ringW, math.Abs(ringF-1)*ringB)
	front := ry > 0
	ringCol := rgb{1, 1, 1}
	ringAlpha := 0.55 * ringCover * (0.45 + 0.55*math.Cos(rx/ringA*math.Pi/2))

	col := bg
	if ringCover > 0 && !front {
		col = over(col, ringCol, ringAlpha)
	}

	// planet sphere
	const R = 0.275
	cx, cy := x-0.5, y-0.5
	dist := math.Hypot(cx, cy)
	if dist <= R {
		nx, ny := cx/R, cy/R
		nz := math.Sqrt(math.Max(0, 1-nx*nx-ny*ny))
		// light from top-left, a bit in front
		l := math.Max(0, nx*-0.5+ny*-0.55+nz*0.67)
		body := mix(rgb{0.28, 0, 0}, rgb{1, 0.17, 0.12}, math.Pow(l, 0.8))
		// polar cap
		if ny < -0.80 {
			body = mix(body, rgb{0.96, 0.93, 0.92}, smooth(-0.80, -0.88, ny)*0.9)
		}
		// canyon scar and craters, in surface coordinates
		canyon := math.Abs(ny-(0.16+0.30*nx+0.04*math.Sin(nx*5))) - 0.035
		if canyon < 0 && math.Abs(nx) < 0.8 {
			body = mix(body, rgb{0.2, 0, 0}, 0.55*smooth(0.035, 0.0, canyon+0.035))
		}
		for _, c := range [][3]float64{{0.50, -0.22, 0.10}, {0.28, -0.42, 0.05}, {-0.52, 0.30, 0.065}} {
			cd := math.Hypot(nx-c[0], ny-c[1])
			if cd < c[2] {
				body = mix(body, rgb{0.3, 0, 0}, 0.55)
			} else if cd < c[2]+0.012 {
				body = mix(body, rgb{1, 0.45, 0.4}, 0.3)
			}
		}
		// night side
		night := smooth(0.35, 1.0, math.Hypot(nx+0.35, ny+0.35)/1.5)
		body = mix(body, rgb{0, 0, 0}, 0.75*night)
		// soft limb
		edge := smooth(R, R-0.004, dist)
		col = mix(col, body, edge)
	} else if dist < R+0.02 {
		col = mix(col, rgb{1, 0.1, 0.05}, 0.25*(1-(dist-R)/0.02))
	}

	if ringCover > 0 && front {
		col = over(col, ringCol, ringAlpha)
	}
	// small satellite on the ring
	sx, sy := 0.30, 0.095
	sdx, sdy := rx-sx, ry-sy
	if sd := math.Hypot(sdx, sdy); sd < 0.022 {
		col = mix(col, rgb{1, 1, 1}, smooth(0.022, 0.016, sd))
	}
	return col, 1
}

func render() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					c, al := shade((float64(px)+(float64(sx)+0.5)/ss)/size, (float64(py)+(float64(sy)+0.5)/ss)/size)
					r += c.r * al
					g += c.g * al
					b += c.b * al
					a += al
				}
			}
			n := float64(ss * ss)
			if a > 0 {
				img.SetNRGBA(px, py, color.NRGBA{uint8(255 * r / a), uint8(255 * g / a), uint8(255 * b / a), uint8(255 * a / n)})
			}
		}
	}
	return img
}

// scale downsamples with a box filter.
func scale(src *image.NRGBA, n int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, n, n))
	f := src.Bounds().Dx() / n
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var r, g, b, a float64
			for j := 0; j < f; j++ {
				for i := 0; i < f; i++ {
					c := src.NRGBAAt(x*f+i, y*f+j)
					al := float64(c.A)
					r += float64(c.R) * al
					g += float64(c.G) * al
					b += float64(c.B) * al
					a += al
				}
			}
			if a > 0 {
				dst.SetNRGBA(x, y, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(a / float64(f*f))})
			}
		}
	}
	return dst
}

func encode(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func writeICO(path string, big *image.NRGBA) {
	sizes := []int{256, 128, 64, 48, 32, 16}
	var pngs [][]byte
	for _, s := range sizes {
		pngs = append(pngs, encode(scale(big, s)))
	}
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, [3]uint16{0, 1, uint16(len(sizes))})
	offset := 6 + 16*len(sizes)
	for i, s := range sizes {
		dim := byte(s)
		if s == 256 {
			dim = 0
		}
		buf.Write([]byte{dim, dim, 0, 0})
		binary.Write(&buf, binary.LittleEndian, [2]uint16{1, 32})
		binary.Write(&buf, binary.LittleEndian, [2]uint32{uint32(len(pngs[i])), uint32(offset)})
		offset += len(pngs[i])
	}
	for _, p := range pngs {
		buf.Write(p)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		panic(err)
	}
}

func main() {
	// 1024 render from a 2048 supersampled pass would be slow; the box filter above is enough at ss=3.
	big := render()
	if err := os.WriteFile("build/appicon.png", encode(big), 0o644); err != nil {
		panic(err)
	}
	writeICO("build/windows/icon.ico", big)
}
