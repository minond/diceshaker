// +build ignore

package main

import (
	"flag"
	"image"
	"image/color"
	"image/jpeg"
	"os"
)

const (
	threshold = 5000
)

var (
	fin  = flag.String("in", "ignore/in.jpg", "Input JPG file")
	fout = flag.String("out", "ignore/out.jpg", "Output JPG file")

	col01 = []uint32{231, 232, 238}    // Original white die
	col02 = []uint32{56, 40, 58}       // Black dot
	col03 = []uint32{0xff, 0xff, 0xff} // Just white
	col04 = []uint32{230, 230, 230}    // White die
	col05 = []uint32{178, 184, 180}    // Darker white die edge
	col06 = []uint32{0x0, 0x0, 0x0}    // Just black
)

func init() {
	flag.Parse()
}

func dist(a1, b1, c1, a2, b2, c2 uint32) uint32 {
	return (a2-a1)*(a2-a1) + (b2-b1)*(b2-b1) + (c2-c1)*(c2-c1)
}

func main() {
	out, err := os.OpenFile(*fout, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}

	file, err := os.Open(*fin)
	if err != nil {
		panic(err)
	}

	img, err := jpeg.Decode(file)
	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	buff := image.NewRGBA(bounds)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			col := img.At(x, y)
			r, g, b, _ := col.RGBA()

			r = r >> 8
			g = g >> 8
			b = b >> 8

			sel := col01
			delta := int(dist(r, g, b, sel[0], sel[1], sel[2]))

			if delta < threshold {
				buff.Set(x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
			}
		}
	}

	if err := jpeg.Encode(out, buff, nil); err != nil {
		panic(err)
	}
}
