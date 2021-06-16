package pkg

import (
	"image"
	"image/color"
)

var _ image.PalettedImage = &Frame{}

// Frame represents a single frame in a DC6.
type Frame struct {
	dc6        *DC6
	Flipped    uint32
	Width      uint32
	Height     uint32
	OffsetX    int32
	OffsetY    int32
	Unknown    uint32
	NextBlock  uint32
	Length     uint32
	FrameData  []byte // size is the value of Length
	Terminator []byte // 3 bytes
	IndexData  []byte
}

func (f *Frame) ColorIndexAt(x, y int) uint8 {
	idx := (y * int(f.Width)) + x

	return f.IndexData[idx]
}

func (f *Frame) ColorModel() color.Model {
	return color.RGBAModel
}

func (f *Frame) Bounds() image.Rectangle {
	origin := image.Point{X: int(f.OffsetX), Y: int(f.OffsetY)}
	delta := image.Point{X: int(f.Width), Y: int(f.Height)}

	return image.Rectangle{
		Min: origin,
		Max: origin.Add(delta),
	}
}

func (f *Frame) At(x, y int) color.Color {
	if f.dc6.palette == nil {
		f.dc6.SetPalette(nil)
	}

	cidx := f.ColorIndexAt(x, y)

	return f.dc6.palette[cidx]
}

func (f *Frame) ToImageRGBA() *image.RGBA {
	img := image.NewRGBA(image.Rectangle{
		Max: f.Bounds().Size(),
	})

	for py := 0; py < int(f.Height); py++ {
		for px := 0; px < int(f.Width); px++ {
			img.Set(px, py, f.At(px, py))
		}
	}

	return img
}
