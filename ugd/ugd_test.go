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
	if len(db["/unitsE/ARMCOM.fbi"]) == 0 {
		t.Error("missing file")
	}
	nodes, err := loadTdfDataDir(db, escUnitsDir)
	if err != nil {
		t.Error(err)
	}
	if len(nodes) == 0 {
		t.Error("got 0 for unit count, expected at least a couple hundred")
	}
	for _, v := range nodes {
		if v.Fields["unitname"] == "CORAK" {
			if v.Fields["copyright"] != "Copyright 1997 Humongous Entertainment. All rights reserved." {
				t.Errorf("%v is missing copyright, has %v", v.Fields["unitname"], v.Fields["copyright"])
			}
		}
	}
	f.Close()
}
