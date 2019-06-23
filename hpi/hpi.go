package hpi

import (
	"io"
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
)

type header struct {
	Marker [4]byte // HAPI
	SaveMarker [4]byte // BANK if a save
	DirectorySize uint32 // The size of the directory
	HeaderKey uint32 // Decryption key
	Start uint32 // File offset of directory
}

type treeInfo struct {
	NumEntries uint32
	EntryOffset uint32
}

func loadHeader(r io.Reader) (h header, err error) {
	err = binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return
	}
	if string(h.Marker[:]) != "HAPI" {
		err = errors.New("invalid file format")
	}
	return
}

func traverseTree(rs io.ReadSeeker, parent string, offset int64) {
	if _, err := rs.Seek(offset, 0); err != nil {
		log.WithFields(log.Fields{
			"parent": parent,
			"offset": offset,
			"err": err,
		}).Fatal("seek failed")
	}
	info := treeInfo{}
	err := binary.Read(rs, binary.LittleEndian, &info)
	if err != nil {
		log.WithFields(log.Fields{
			"parent": parent,
			"offset": offset,
			"err": err,
		}).Fatal("treeInfo read failed")
	}
}






