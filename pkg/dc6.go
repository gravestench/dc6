package pkg

import (
	"fmt"
	"image/color"
	"math"

	"github.com/gravestench/bitstream"
)

const (
	endOfScanLine = 0x80
	maxRunLength  = 0x7f
)

type scanlineState int

const (
	endOfLine scanlineState = iota
	runOfTransparentPixels
	runOfOpaquePixels
)

// DC6 represents a DC6 file.
type DC6 struct {
	Version     int32
	Flags       uint32
	Encoding    uint32
	Termination []byte // 4 bytes
	Directions  []*Direction
	palette     color.Palette
}

type Direction struct {
	Frames []*Frame // size is Directions*FramesPerDirection
}

// FromBytes uses restruct to read the binary dc6 data into structs then parses image data from the frame data.
func FromBytes(data []byte) (result *DC6, err error) {
	result = &DC6{}

	stream := bitstream.NewReader().FromBytes(data...)

	if err = result.decodeHeader(stream); err != nil {
		return nil, err
	}

	if err = result.decodeBody(stream); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DC6) decodeHeader(stream *bitstream.Reader) (err error) {
	const (
		versionBytes     = 4
		flagsBytes       = 4
		encodingBytes    = 4
		terminationBytes = 4
	)

	// only check last err
	d.Version, _ = stream.Next(versionBytes).Bytes().AsInt32()
	d.Flags, _ = stream.Next(flagsBytes).Bytes().AsUInt32()
	d.Encoding, _ = stream.Next(encodingBytes).Bytes().AsUInt32()
	d.Termination, err = stream.Next(terminationBytes).Bytes().AsBytes()

	return err
}

func (d *DC6) decodeBody(stream *bitstream.Reader) (err error) {
	const (
		terminatorSize          = 3
		directionsBytes         = 4
		framesPerDirectionBytes = 4
		framePointerBytes       = 4
	)

	const (
		frameFlippedBytes   = 4
		frameWidthBytes     = 4
		frameHeightBytes    = 4
		frameOffsetXBytes   = 4
		frameOffsetYBytes   = 4
		frameUnknownBytes   = 4
		frameNextBlockBytes = 4
		frameLengthBytes    = 4
	)

	numDirections, _ := stream.Next(directionsBytes).Bytes().AsUInt32()
	framesPerDirection, err := stream.Next(framesPerDirectionBytes).Bytes().AsUInt32()
	totalFrames := int(numDirections * framesPerDirection)

	d.Directions = make([]*Direction, numDirections)

	for i := 0; i < totalFrames; i++ {
		if _, err = stream.Next(framePointerBytes).Bytes().AsUInt32(); err != nil {
			return err
		}
	}

	for idx := 0; idx < totalFrames; idx++ {
		dirIdx := idx / int(framesPerDirection)
		frameIdx := idx % int(framesPerDirection)

		if d.Directions[dirIdx] == nil {
			d.Directions[dirIdx] = &Direction{}
		}

		if d.Directions[dirIdx].Frames == nil {
			d.Directions[dirIdx].Frames = make([]*Frame, framesPerDirection)
		}

		frame := &Frame{dc6: d}

		// toss the errors, only check last err
		frame.Flipped, _ = stream.Next(frameFlippedBytes).Bytes().AsUInt32()
		frame.Width, _ = stream.Next(frameWidthBytes).Bytes().AsUInt32()
		frame.Height, _ = stream.Next(frameHeightBytes).Bytes().AsUInt32()
		frame.OffsetX, _ = stream.Next(frameOffsetXBytes).Bytes().AsInt32()
		frame.OffsetY, _ = stream.Next(frameOffsetYBytes).Bytes().AsInt32()
		frame.Unknown, _ = stream.Next(frameUnknownBytes).Bytes().AsUInt32()
		frame.NextBlock, _ = stream.Next(frameNextBlockBytes).Bytes().AsUInt32()
		frame.Length, _ = stream.Next(frameLengthBytes).Bytes().AsUInt32()
		frame.FrameData, _ = stream.Next(int(frame.Length)).Bytes().AsBytes()
		frame.Terminator, err = stream.Next(terminatorSize).Bytes().AsBytes()

		if err != nil {
			return fmt.Errorf("could not decode body, %w", err)
		}

		d.Directions[dirIdx].Frames[frameIdx] = frame
	}

	for idx := range d.Directions {
		d.Directions[idx].decodeFrames()
	}

	return nil
}

// decodeFrame decodes the given frame to an indexed color texture
func (d *Direction) decodeFrames() {
	for idx := range d.Frames {
		d.decodeFrame(idx)
	}
}

func (d *Direction) decodeFrame(frameIndex int) {
	frame := d.Frames[frameIndex]

	indexData := make([]byte, frame.Width*frame.Height)
	x := 0
	y := int(frame.Height) - 1
	offset := 0

loop: // this is a label for the loop, so the switch can break the loop (and not the switch)
	for {
		if offset >= len(frame.FrameData) {
			break
		}

		b := int(frame.FrameData[offset])
		offset++

		switch scanlineType(b) {
		case endOfLine:
			if y == 0 {
				break loop
			}

			y--

			x = 0
		case runOfTransparentPixels:
			transparentPixels := b & maxRunLength
			x += transparentPixels
		case runOfOpaquePixels:
			for i := 0; i < b; i++ {
				index := x + y*int(frame.Width) + i
				if index < len(indexData) && offset < len(frame.FrameData) {
					indexData[index] = frame.FrameData[offset]
				}
				offset++
			}

			x += b
		}
	}

	frame.IndexData = indexData
}

func scanlineType(b int) scanlineState {
	if b == endOfScanLine {
		return endOfLine
	}

	if (b & endOfScanLine) > 0 {
		return runOfTransparentPixels
	}

	return runOfOpaquePixels
}

// Clone creates a copy of the DC6
func (d *DC6) Clone() *DC6 {
	clone := *d

	copy(clone.Termination, d.Termination)

	clone.Directions = make([]*Direction, len(d.Directions))
	for dirIdx := range d.Directions {
		clone.Directions[dirIdx].Frames = make([]*Frame, len(d.Directions[dirIdx].Frames))
		for frameIdx := range d.Directions[dirIdx].Frames {
			frame := *d.Directions[dirIdx].Frames[frameIdx]
			clone.Directions[dirIdx].Frames[frameIdx] = &frame
		}
	}

	return &clone
}

// Palette returns the current color palette
func (d *DC6) Palette() color.Palette {
	if d.palette == nil {
		d.SetPalette(nil)
	}

	return d.palette
}

// SetPalette sets the current color palette
func (d *DC6) SetPalette(p color.Palette) {
	if p == nil {
		p = d.getDefaultPalette()
	}

	d.palette = p
}

func (d *DC6) getDefaultPalette() color.Palette {
	const numColors = 256

	palette := make(color.Palette, numColors)

	for idx := range palette {
		rgb := uint8(idx)
		c := color.RGBA{}
		c.R, c.G, c.B, c.A = rgb, rgb, rgb, math.MaxUint8

		palette[idx] = c
	}

	return palette
}
