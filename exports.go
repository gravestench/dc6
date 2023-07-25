package dc6

import (
	"github.com/gravestench/dc6/pkg"
)

// just a handful of aliases to handle importing from repo root

type (
	DC6         = pkg.DC6
	Header      = pkg.Header
	Direction   = pkg.Direction
	Frame       = pkg.Frame
	FrameHeader = pkg.FrameHeader
)

func FromBytes(data []byte) (result *DC6, err error) {
	return pkg.FromBytes(data)
}
