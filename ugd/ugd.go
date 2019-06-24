package ugd

import (
	"github.com/cosmouser/tdf"
	"path"
	"strings"
	"bytes"
)

const (
	escUnitsDir = "unitsE"
	escWeaponsDir = "weaponE"
	escUnitpicsDir = "unitpicE"
	escDownloadsDir = "downloadsE"
)

func loadTdfDataDir(store map[string][]byte, dir string) (nodes []*tdf.Node, err error) {
	fileNames := []string{}
	dirPath := path.Join("/", dir)
	for k := range store {
		if strings.Index(k, dirPath) == 0 {
			fileNames = append(fileNames, k)
		}
	}
	for _, fileName := range fileNames {
		if ext := strings.ToLower(path.Ext(fileName)); ext == ".tdf" || ext == ".ota" || ext == ".fbi" {
			tmp, err := tdf.Decode(bytes.NewReader(store[fileName]))
			if err != nil {
				return nodes, err
			}
			for _, v := range tmp {
				nodes = append(nodes, v)
			}
		}
	}
	return
}

