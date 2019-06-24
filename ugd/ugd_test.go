package ugd

import (
	"testing"
	"os"
	"strings"
	"path"
	"encoding/gob"
)

func openGob() (store map[string][]byte, err error) {
	f, err := os.Open(path.Join("..", "TAESC.gp3.gob"))
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(f)
	err = dec.Decode(&store)
	if err != nil {
		return nil, err
	}
	f.Close()
	return
}

func TestMakeUnitRecords(t *testing.T) {
	db, err := openGob()
	if err != nil {
		t.Error(err)
	}
	nodes, err := loadTdfDataDir(db, escUnitsDir)
	if err != nil {
		t.Error(err)
	}
	downloadNodes, err := loadTdfDataDir(db, escDownloadsDir)
	if err != nil {
		t.Error(err)
	}
	addBuildRelationships(nodes, downloadNodes)
	records, err := makeUnitRecords(nodes)
	if err != nil {
		t.Error(err)
	}
	if len(records) < 1 {
		t.Fatal("no records were created")
	}
	if len(records[0]) < 2 {
		t.Error("no fields were created")
	}
}




func TestUnitData(t *testing.T) {
	db, err := openGob()
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
	downloadNodes, err := loadTdfDataDir(db, escDownloadsDir)
	if err != nil {
		t.Error(err)
	}
	addBuildRelationships(nodes, downloadNodes)
	for _, v := range nodes {
		if v.Fields["unitname"] == "CORAK" {
			if v.Fields["copyright"] != "Copyright 1997 Humongous Entertainment. All rights reserved." {
				t.Errorf("%v is missing copyright, has %v", v.Fields["unitname"], v.Fields["copyright"])
			}
			if strings.Index(v.Fields["builtby"], "CORLAB") < 0 {
				t.Error("builtby info for CORAK is missing CORLAB")
			}
		}
		if v.Fields["unitname"] == "CORCK" {
			if strings.Index(v.Fields["canbuild"], "CORRL") < 0 {
				t.Error("canbuild info for CORCK is missing CORRL")
			}
		}
	}
}