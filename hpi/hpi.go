package hpi

import (
	"io"
	"encoding/binary"
)

type header struct {
	Marker [4]byte // HAPI
	SaveMarker [4]byte // BANK if a save
	DirectorySize uint32 // The size of the directory
	HeaderKey uint32 // Decryption key
	Start uint32 // File offset of directory
}

func loadHeader(r io.Reader) (h header, err error) {
	err = binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return
	}
	return
}




