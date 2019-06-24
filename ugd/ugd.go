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

func addBuildRelationships(unitinfoList, buildinfoList []*tdf.Node) {
	unitsIndex := make(map[string]int)
	for i, v := range unitinfoList {
		unitsIndex[v.Fields["unitname"]] = i
	}
	for _, v := range buildinfoList {
		if str, ok := v.Fields["unitmenu"]; ok {
			if unitIndex, ok := unitsIndex[str]; ok {
				canbuildList := unitinfoList[unitIndex].Fields["canbuild"]
				if strings.Index(canbuildList, v.Fields["unitname"]) < 0 {
					if canbuildList != "" {
						unitinfoList[unitIndex].Fields["canbuild"] += " " + v.Fields["unitname"]
					} else {
						unitinfoList[unitIndex].Fields["canbuild"] += v.Fields["unitname"]
					}
				}
			}
		}
		if str, ok := v.Fields["unitname"]; ok {
			if unitIndex, ok := unitsIndex[str]; ok {
				builtbyList := unitinfoList[unitIndex].Fields["builtby"]
				if strings.Index(builtbyList, v.Fields["unitmenu"]) < 0 {
					if builtbyList != "" {
						unitinfoList[unitIndex].Fields["builtby"] += " " + v.Fields["unitmenu"]
					} else {
						unitinfoList[unitIndex].Fields["builtby"] += v.Fields["unitmenu"]
					}
				}
			}
		}
	}
}

func makeUnitRecords(unitinfoList []*tdf.Node) (records [][]string, err error) {
	// Ensure unitname is the first field.
	fields := []string{"unitname"}
	fieldsMap := make(map[string]int)
	fieldsMap["unitname"] = 1
	fieldCursor := 1
	for _, unitNode := range unitinfoList {
		for k := range unitNode.Fields {
			if fieldNumber, ok := fieldsMap[k]; !ok && fieldNumber == 0 {
				fields = append(fields, k)
				fieldCursor++
				fieldsMap[k] = fieldCursor
			}
		}
	}
	records = append(records, fields)
	for _, unitNode := range unitinfoList {
		tmp := make([]string, len(fields))
		for k, v := range unitNode.Fields {
			tmp[fieldsMap[k]-1] = v
		}
		records = append(records, tmp)
	}
	return
}

func makeWeaponRecords(weaponList []*tdf.Node) (records [][]string, err error) {
	// Ensure weapname is the first field.
	fields := []string{"weapname"}
	fieldsMap := make(map[string]int)
	fieldsMap["weapname"] = 1
	fieldCursor := 1
	for _, weapNode := range weaponList {
		for k := range weapNode.Fields {
			if fieldNumber, ok := fieldsMap[k]; !ok && fieldNumber == 0 {
				fields = append(fields, k)
				fieldCursor++
				fieldsMap[k] = fieldCursor
			}
		}
	}
	for _, weapNode := range weaponList {
		for k := range weapNode.Children[0].Fields {
			damageKey := weapNode.Children[0].Name + "_" + k
			if fieldNumber, ok := fieldsMap[damageKey]; !ok && fieldNumber == 0 {
				fields = append(fields, damageKey)
				fieldCursor++
				fieldsMap[damageKey] = fieldCursor
			}
		}
	}
	records = append(records, fields)
	return
}


