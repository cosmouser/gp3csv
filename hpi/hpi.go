package hpi

import (
	"io"
	"encoding/binary"
	"errors"
	"path"
	"bufio"
	log "github.com/sirupsen/logrus"
)
const (
	dirEntryLength = 9
)

type header struct {
	Marker [4]byte // HAPI
	SaveMarker [4]byte // BANK if a save
	DirectorySize uint32 // The size of the directory
	HeaderKey uint32 // Decryption key
	Start uint32 // File offset of directory
}

type dirInfo struct {
	NumEntries uint32
	EntryOffset uint32
}

type dirEntry struct {
	NameOffset uint32 // Points to the filename
	DirDataOffset uint32 // Points to the directory data
	Flag byte // If this is 1 the entry is a subdirectory and if it is 0 it is a file
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

func traverseTree(rs io.ReadSeeker, parent string, offset int64, store map[string][]byte) {
	if n, err := rs.Seek(offset, io.SeekStart); err != nil {
		log.WithFields(log.Fields{
			"parent": parent,
			"offset": offset,
			"newpos": n,
			"err": err,
		}).Fatal("seek failed")
	}
	info := dirInfo{}
	err := binary.Read(rs, binary.LittleEndian, &info)
	if err != nil {
		log.WithFields(log.Fields{
			"parent": parent,
			"offset": offset,
			"err": err,
		}).Fatal("dirInfo read failed")
	}
	for i := 0; i < int(info.NumEntries); i++ {
		dirEntryPointer := int64(info.EntryOffset)+(int64(i)*dirEntryLength)
		if n, err := rs.Seek(dirEntryPointer, io.SeekStart); err != nil {
			log.WithFields(log.Fields{
				"parent": parent,
				"dirEntryPointer": dirEntryPointer,
				"newpos": n,
				"err": err,
			}).Fatal("seek failed")
		}
		entry := dirEntry{}
		err = binary.Read(rs, binary.LittleEndian, &entry)
		if n, err := rs.Seek(int64(entry.NameOffset), io.SeekStart); err != nil {
			log.WithFields(log.Fields{
				"parent": parent,
				"entry.NameOffset": entry.NameOffset,
				"newpos": n,
				"err": err,
			}).Fatal("seek failed")
		}
		nameReader := bufio.NewReader(rs)
		fileName, err := nameReader.ReadBytes('\x00')
		if err != nil {
			log.WithFields(log.Fields{
				"parent": parent,
				"entry.NameOffset": entry.NameOffset,
				"fileName": fileName,
				"err": err,
			}).Fatal("nameReader failed")
		}
		name := path.Join(parent, string(fileName[:len(fileName)-1]))
		if entry.Flag == 1 {
			// dirEntry is for directory, recurse
			traverseTree(rs, name, int64(entry.DirDataOffset), store)
		} else {
			// dirEntry is for file
			store[name] = []byte{}
		}
	}
}
