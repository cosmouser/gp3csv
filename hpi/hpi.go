package hpi

type header struct {
	marker []byte // HAPI
	saveMaker []byte // BANK if a save
	directorySize uint32 // The size of the directory
	headerKey uint32 // Decryption key
	start uint32 // File offset of directory
}
