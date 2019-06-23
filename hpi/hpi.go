package hpi

import (
	"io"
	"fmt"
	"encoding/binary"
)

type header struct {
	marker [4]byte // HAPI
	saveMarker [4]byte // BANK if a save
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
	if n, err := r.Read(h.saveMarker[:]); err != nil || n != len(h.saveMarker) {
		if err != nil {
			return h, err
		}
		return h, fmt.Errorf("expecting to read %v bytes, only read %v bytes", len(h.saveMarker), n)
	}
	err = binary.Read(r, binary.LittleEndian, &h.directorySize)
	if err != nil {
		return
	}
	err = binary.Read(r, binary.LittleEndian, &h.headerKey)
	if err != nil {
		return
	}
	err = binary.Read(r, binary.LittleEndian, &h.start)
	if err != nil {
		return
	}
	return
}




