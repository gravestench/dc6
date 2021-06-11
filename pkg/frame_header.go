package pkg

// FrameHeader represents the header of a frame in a DC6.
type FrameHeader struct {
	Flipped   int32  `struct:"int32"`
	Width     int32  `struct:"int32"`
	Height    int32  `struct:"int32"`
	OffsetX   int32  `struct:"int32"`
	OffsetY   int32  `struct:"int32"`
	Unknown   uint32 `struct:"uint32"`
	NextBlock uint32 `struct:"uint32"`
	Length    uint32 `struct:"uint32"`
}
