package pkg

import (
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
	stream             *bitstream.Reader
	Version            int32
	Flags              uint32
	Encoding           uint32
	Termination        []byte // 4 bytes
	Directions         uint32
	FramesPerDirection uint32
	FramePointers      []uint32 // size is Directions*FramesPerDirection
	Frames             []*Frame // size is Directions*FramesPerDirection
}

// FromBytes uses restruct to read the binary dc6 data into structs then parses image data from the frame data.
func FromBytes(data []byte) (result *DC6, err error) {
	result = &DC6{
		stream: bitstream.NewReader().FromBytes(data...),
	}

	if err = result.decodeHeader(); err != nil {
		return nil, err
	}

	if err = result.decodeBody(); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *DC6) decodeHeader() (err error) {
	const terminationSize = 4

	// only check last err
	d.Version, _ = d.stream.Next(4).Bytes().AsInt32()
	d.Flags, _ = d.stream.Next(4).Bytes().AsUInt32()
	d.Encoding, _ = d.stream.Next(4).Bytes().AsUInt32()
	d.Termination, _ = d.stream.Next(terminationSize).Bytes().AsBytes()
	d.Directions, _ = d.stream.Next(4).Bytes().AsUInt32()
	d.FramesPerDirection, err = d.stream.Next(4).Bytes().AsUInt32()

	if err != nil {
		return err
	}

	frameCount := int(d.Directions * d.FramesPerDirection)

	d.FramePointers = make([]uint32, frameCount)
	for i := 0; i < frameCount; i++ {
		if d.FramePointers[i], err = d.stream.Next(4).Bytes().AsUInt32(); err != nil {
			return err
		}
	}

	return nil
}

func (d *DC6) decodeBody() (err error) {
	const terminatorSize  = 3

	frameCount := int(d.Directions * d.FramesPerDirection)

	d.Frames = make([]*Frame, frameCount)

	for i := 0; i < frameCount; i++ {
		frame := &Frame{}

		// toss the errors, only check last err
		frame.Flipped, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.Width, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.Height, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.OffsetX, _ = d.stream.Next(4).Bytes().AsInt32()
		frame.OffsetY, _ = d.stream.Next(4).Bytes().AsInt32()
		frame.Unknown, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.NextBlock, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.Length, _ = d.stream.Next(4).Bytes().AsUInt32()
		frame.FrameData, _ = d.stream.Next(int(frame.Length)).Bytes().AsBytes()
		frame.Terminator, err = d.stream.Next(terminatorSize).Bytes().AsBytes()

		if err != nil {
			return err
		}

		d.Frames[i] = frame
	}

	return nil
}

// DecodeFrame decodes the given frame to an indexed color texture
func (d *DC6) DecodeFrame(frameIndex int) []byte {
	frame := d.Frames[frameIndex]

	indexData := make([]byte, frame.Width*frame.Height)
	x := 0
	y := int(frame.Height) - 1
	offset := 0

loop: // this is a label for the loop, so the switch can break the loop (and not the switch)
	for {
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
				indexData[x+y*int(frame.Width)+i] = frame.FrameData[offset]
				offset++
			}

			x += b
		}
	}

	return indexData
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
	copy(clone.FramePointers, d.FramePointers)
	clone.Frames = make([]*Frame, len(d.Frames))

	for i := range d.Frames {
		cloneFrame := *d.Frames[i]
		clone.Frames = append(clone.Frames, &cloneFrame)
	}

	return &clone
}
