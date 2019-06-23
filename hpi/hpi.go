package hpi

import (
	"io"
	"fmt"
)

type header struct {
	marker [4]byte // HAPI
	saveMaker [4]byte // BANK if a save
	directorySize uint32 // The size of the directory
	headerKey uint32 // Decryption key
	start uint32 // File offset of directory
}

func loadHeader(r io.Reader) (h header, err error) {
	if n, err := r.Read(h.marker[:]); err != nil || n != len(h.marker) {
		if err != nil {
			return h, err
		}
		return h, fmt.Errorf("expecting to read %v bytes, only read %v bytes", len(h.marker), n)
	}
	return
}




