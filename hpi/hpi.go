package hpi

import (
	"strings"
	"io"
	"encoding/binary"
	"bytes"
	"compress/zlib"
	"errors"
	"path"
	"bufio"
	"math"
	log "github.com/sirupsen/logrus"
)
const (
	dirEntryLength = 9
	maxChunkSize = 65536
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

type fileData struct {
	DataOffset uint32 // Starting offset of the file
	FileSize uint32 // Size of the decompressed file
	Flag byte // 1: compressed with LZ77, 2: compressed iwth zlib, 0: not compressed
}

type dirEntry struct {
	NameOffset uint32 // Points to the filename
	DirDataOffset uint32 // Points to the directory data
	Flag byte // If this is 1 the entry is a subdirectory and if it is 0 it is a file
}

type chunk struct {
	_ byte
	CompMethod byte // 1=LZ77, 2=zlib
	Encrypt byte // Is the block encrypted?
	CompressedSize uint32 // Length of the compressed data
	DecompressedSize uint32 // Length of the decompressed data
	Checksum uint32 // A sum of all the bytes of the data
}

func decryptChunkData(ec []byte, compressedSize int) (dc []byte) {
	dc = make([]byte, compressedSize)
	for i := 0; i < compressedSize; i++ {
		dc[i] = (ec[i] - byte(i)) ^ byte(i)
	}
	return
}

func processFile(r io.Reader, numChunks int) (out []byte, err error) {
	var buf, input io.Reader
	var writer bytes.Buffer
	markerBuf := make([]byte, 4)
	for i := 0; i < numChunks; i++ {
		if n, err := r.Read(markerBuf); err != nil || n != len(markerBuf) {
			log.WithFields(log.Fields{
				"numChunks": numChunks,
				"i": i,
			}).Fatal("chunk marker read failed")
		}
		for string(markerBuf) != "SQSH" {
			if n, err := r.Read(markerBuf); err != nil || n != len(markerBuf) {
				log.WithFields(log.Fields{
					"numChunks": numChunks,
					"i": i,
				}).Fatal("chunk marker read failed")
			}
		}
		c := chunk{}
		err = binary.Read(r, binary.LittleEndian, &c)
		chunkData := make([]byte, int(c.CompressedSize))
		if n, err := r.Read(chunkData); err != nil || n != len(chunkData) {
			return out, err
		}
		if c.Encrypt == 1 {
			buf = bytes.NewReader(decryptChunkData(chunkData, int(c.CompressedSize)))
		} else {
			buf = bytes.NewReader(chunkData)
		}
		switch method := c.CompMethod; method {
		case 0:
			input = buf
		case 1:
			log.Fatal("chunk uses LZ77 compression, not implemented")
		case 2:
			input, err = zlib.NewReader(buf)
			if err != nil {
				return
			}
		}
		io.Copy(&writer, input)
	}
	out = writer.Bytes()
	return
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
		name := strings.ToLower(path.Join(parent, string(fileName[:len(fileName)-1])))
		if entry.Flag == 1 {
			// dirEntry is for directory, recurse
			traverseTree(rs, name, int64(entry.DirDataOffset), store)
		} else {
			// dirEntry is for file
			if n, err := rs.Seek(int64(entry.DirDataOffset), io.SeekStart); err != nil {
				log.WithFields(log.Fields{
					"name": name,
					"entry.DirDataOffset": entry.DirDataOffset,
					"newpos": n,
					"err": err,
				}).Fatal("seek failed")
			}
			fd := fileData{}
			err = binary.Read(rs, binary.LittleEndian, &fd)
			if err != nil {
				log.WithFields(log.Fields{
					"name": name,
					"err": err,
				}).Fatal("fileData read failed")
			}
			chunks := int(math.Ceil(float64(fd.FileSize) / maxChunkSize))
			if int(fd.FileSize)%maxChunkSize == 0 {
				chunks++
			}
			if n, err := rs.Seek(int64(fd.DataOffset), io.SeekStart); err != nil {
				log.WithFields(log.Fields{
					"name": name,
					"fd.DataOffset": fd.DataOffset,
					"newpos": n,
					"err": err,
				}).Fatal("seek failed")
			}
			// Process the file and load it into the store
			outData, err := processFile(rs, chunks)
			if err != nil {
				log.WithFields(log.Fields{
					"name": name,
					"err": err,
				}).Fatal("processFile failed")
			}
			store[name] = outData

		}
	}
}
