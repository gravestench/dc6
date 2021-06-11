package pkg

// Frame represents a single frame in a DC6.
type Frame struct {
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
}
