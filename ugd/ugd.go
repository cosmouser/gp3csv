package ugd

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/cosmouser/pcx"
	"github.com/cosmouser/tdf"
)

var (
	unitsDir     = "unitsE"
	weaponDir    = "weaponE"
	unitpicsDir  = "unitpicE"
	downloadsDir = "downloadsE"
)

// SetPaths sets custom paths
func SetPaths(units, weapon, unitpics, downloads string) {
	unitsDir = units
	weaponDir = weapon
	unitpicsDir = unitpics
	downloadsDir = downloads
}

// Manifest has information for TADA on how to handle the mod data
type Manifest struct {
	ModName   string   `json:"modName"`
	Checksum  string   `json:"checksum"`
	Extension string   `json:"ext"`
	Features  []string `json:"features"`
}

// CreateModCassette extracts data for TADA
func CreateModCassette(store map[string][]byte, info Manifest, out io.Writer) (err error) {
	unitNodes, err := loadTdfDataDir(store, unitsDir)
	if err != nil {
		return err
	}
	// Create a zip with a manifest.json, units json and unitpic directory of images
	// The units json is an alphabetically sorted array by unitname
	var (
		unitsInfo []map[string]string
		unitsTmp  map[int]map[string]string
	)
	unitnames := []string{}
	unitsIndex := make(map[string]int)
	unitsTmp = make(map[int]map[string]string)
	for _, v := range unitNodes {
		unitnames = append(unitnames, v.Fields["unitname"])
	}
	sort.Strings(unitnames)
	for i, v := range unitnames {
		unitsIndex[v] = i
	}

	for _, v := range unitNodes {
		tmp := make(map[string]string)
		for _, f := range info.Features {
			feature := strings.ToLower(f)
			tmp[feature] = v.Fields[feature]
		}
		unitsTmp[unitsIndex[v.Fields["unitname"]]] = tmp
	}
	unitsInfo = make([]map[string]string, len(unitNodes))
	for i := 0; i < len(unitsInfo); i++ {
		unitsInfo[i] = unitsTmp[i]
	}

	zipWriter := zip.NewWriter(out)
	f, err := zipWriter.Create(path.Join(info.Checksum, "manifest.json"))
	if err != nil {
		return err
	}
	dat, err := json.Marshal(info)
	if err != nil {
		return err
	}
	f.Write(dat)
	f, err = zipWriter.Create(path.Join(info.Checksum, "units.json"))
	if err != nil {
		return err
	}
	dat, err = json.Marshal(unitsInfo)
	if err != nil {
		return err
	}
	f.Write(dat)

	pics, err := GatherUnitPics(store)
	if err != nil {
		return err
	}
	var imageBuf bytes.Buffer
	for k, v := range pics {
		f, err := zipWriter.Create(path.Join(info.Checksum, "unitpic", strings.ToLower(k)))
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
	return err
}

// EncodeUnitsCSV writes tdf unit data in CSV format
func EncodeUnitsCSV(store map[string][]byte, out io.Writer) (err error) {
	unitNodes, err := loadTdfDataDir(store, unitsDir)
	if err != nil {
		return err
	}
	downloadNodes, err := loadTdfDataDir(store, downloadsDir)
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
	weapNodes, err := loadTdfDataDir(store, weaponDir)
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
	dirPath := path.Join("/", unitpicsDir)
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
