package pkg

// Header represents the file header of a DC6 file.
type Header struct {
	Version            int32  `struct:"int32"`
	Flags              uint32 `struct:"uint32"`
	Encoding           uint32 `struct:"uint32"`
	Termination        []byte `struct:"[4]byte"`
	Directions         int32  `struct:"int32"`
	FramesPerDirection int32  `struct:"int32"`
}
