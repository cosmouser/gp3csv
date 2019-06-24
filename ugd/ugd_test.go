package ugd

import (
	"testing"
	"os"
	"path"
	"encoding/gob"
)

func TestPrintUnit(t *testing.T) {
	f, err := os.Open(path.Join("..", "TAESC.gp3.gob"))
	if err != nil {
		t.Error(err)
	}
	dec := gob.NewDecoder(f)
	db := make(map[string][]byte)
	err = dec.Decode(&db)
	if err != nil {
		t.Error(err)
	}
	f.Close()
}
