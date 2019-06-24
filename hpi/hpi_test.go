package hpi

import (
	"testing"
	"path"
	"os"
)

func TestLoadHeader(t *testing.T) {
	file, err := os.Open(path.Join("..", "TAESC.gp3"))
	if err != nil {
		t.Error(err)
	}
	header, err := loadHeader(file)
	if err != nil {
		t.Error(err)
	}
	if string(header.Marker[:]) != "HAPI" {
		t.Errorf("expected HAPI, got %v", header.Marker)
	}
	if header.DirectorySize == 0 {
		t.Error("got zero value for header.DirectorySize")
	}
	if header.Start == 0 {
		t.Error("got zero value for header.Start")
	}
	file.Close()
}

func TestTraverseTree(t *testing.T) {
	file, err := os.Open(path.Join("..", "TAESC.gp3"))
	if err != nil {
		t.Error(err)
	}
	h, err := loadHeader(file)
	if err != nil {
		t.Error(err)
	}
	// add all of the filenames to a map and check the size of the map
	result := make(map[string][]byte)
	traverseTree(file, "/", int64(h.Start), result)
	if len(result) == 0 {
		t.Error("traverseTree failed to traverse the archive")
	}
	if len(result["/unitsE/ARMCOM.FBI"]) == 0 {
		t.Error("got zero value for size of /unitsE/ARMCOM.FBI")
	}
	file.Close()
}


