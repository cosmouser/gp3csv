package ugd

import (
	"github.com/cosmouser/tdf"
	"github.com/cosmouser/pcx"
	"image"
	"image/png"
	"io"
	"archive/zip"
	"path"
	"strings"
	"bytes"
	"encoding/csv"
)

const (
	escUnitsDir = "unitsE"
	escWeaponsDir = "weaponE"
	escUnitpicsDir = "unitpicE"
	escDownloadsDir = "downloadsE"
)

// EncodeUnitsCSV writes tdf unit data in CSV format
func EncodeUnitsCSV(store map[string][]byte, out io.Writer) (err error) {
	unitNodes, err := loadTdfDataDir(store, escUnitsDir)
	if err != nil {
		return err
	}
	downloadNodes, err := loadTdfDataDir(store, escDownloadsDir)
	if err != nil {
		return err
	}
	addBuildRelationships(unitNodes, downloadNodes)
	unitRecords, err := makeUnitRecords(unitNodes)
	if err != nil {
		return err
	}
	csvWriter := csv.NewWriter(out)
	err = csvWriter.WriteAll(unitRecords)
	csvWriter.Flush()
	return
}

// EncodeWeaponsCSV writes tdf weapon data in CSV format
func EncodeWeaponsCSV(store map[string][]byte, out io.Writer) (err error) {
	weapNodes, err := loadTdfDataDir(store, escWeaponsDir)
	if err != nil {
		return err
	}
	weapRecords, err := makeWeaponRecords(weapNodes)
	if err != nil {
		return err
	}
	csvWriter := csv.NewWriter(out)
	err = csvWriter.WriteAll(weapRecords)
	csvWriter.Flush()
	return
}
// ExportPicsToZip writes the unitpics to a zip file
func ExportPicsToZip(pics map[string]image.Image, out io.Writer) (err error) {
	zipWriter := zip.NewWriter(out)
	var imageBuf bytes.Buffer
	for k, v := range pics {
		f, err := zipWriter.Create(k)
		if err != nil {
			return err
		}
		err = png.Encode(&imageBuf, v)
		if err != nil {
			return err
		}
		f.Write(imageBuf.Bytes())
		imageBuf.Reset()
	}
	err = zipWriter.Close()
	return
}
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
// GatherUnitPics extracts the pcx unit pics from the memory store
func GatherUnitPics(store map[string][]byte) (pics map[string]image.Image, err error) {
	pics = make(map[string]image.Image)
	fileNames := []string{}
	dirPath := path.Join("/", escUnitpicsDir)
	for k := range store {
		if strings.Index(k, dirPath) == 0 {
			fileNames = append(fileNames, k)
		}
	}
	for _, fileName := range fileNames {
		if ext := strings.ToLower(path.Ext(fileName)); ext == ".pcx" {
			tmp, err := pcx.Decode8Bit256Color(bytes.NewReader(store[fileName]))
			if err != nil {
				return pics, err
			}
			pics[path.Base(fileName)+".png"] = tmp
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
	for _, weapNode := range weaponList {
		tmp := make([]string, len(fields))
		tmp[0] = weapNode.Name
		for k, v := range weapNode.Fields {
			tmp[fieldsMap[k]-1] = v
		}
		for k, v := range weapNode.Children[0].Fields {
			damageKey := weapNode.Children[0].Name + "_" + k
			tmp[fieldsMap[damageKey]-1] = v
		}
		records = append(records, tmp)
	}
	return
}


