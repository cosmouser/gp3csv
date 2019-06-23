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
	header, err := scanHeader(file)
	if err != nil {
		t.Error(err)
	}
	if header.marker != "HAPI" {
		t.Error("expected HAPI, got %v", header.marker)
	}
	file.Close()
}

