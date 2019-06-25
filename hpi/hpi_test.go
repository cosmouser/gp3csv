package hpi

import (
	"testing"
	"encoding/gob"
	"time"
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
	t0 := time.Now()
	traverseTree(file, "/", int64(h.Start), result)
	t.Logf("traverseTree took %v", time.Since(t0))
	if len(result) == 0 {
		t.Error("traverseTree failed to traverse the archive")
	}
	if len(result["/unitsE/ARMCOM.fbi"]) == 0 {
		t.Error("got zero value for size of /unitsE/ARMCOM.fbi")
	}
	file.Close()
	if os.Getenv("HPI2GOB") == "GO" {
		out, err := os.Create(path.Join("..", "TAESC.gp3.gob"))
		if err != nil {
			t.Error(err)
		}
		enc := gob.NewEncoder(out)
		err = enc.Encode(result)
		if err != nil {
			t.Error(err)
		}
		out.Close()
	}
}

