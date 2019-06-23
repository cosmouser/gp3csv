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
	if string(header.marker[:]) != "HAPI" {
		t.Errorf("expected HAPI, got %v", header.marker)
	}
	if header.directorySize == 0 {
		t.Error("got zero value for header.directorySize")
	}
	if header.start == 0 {
		t.Error("got zero value for header.start")
	}
	file.Close()
}

