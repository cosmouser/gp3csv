package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cosmouser/gp3csv/hpi"
	"github.com/cosmouser/gp3csv/ugd"
)

const (
	unitsDir     = "units"
	weaponDir    = "weapon"
	unitpicDir   = "unitpic"
	downloadsDir = "downloads"
)

func main() {
	unitsBool := flag.Bool("u", true, "extract unit data")
	weaponsBool := flag.Bool("w", true, "extract weapon data")
	imgBool := flag.Bool("p", true, "extract unitpics")
	tadaBool := flag.Bool("t", false, "extract mod data for tada")
	tadaOpts := flag.String("topts", "", "modName,first4Checksum,demoExt,feature1 feature2")
	versBool := flag.Bool("v", false, "prints the version and exits")
	flag.Parse()
	if *versBool {
		fmt.Println("gp3csv version 1.1.0")
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: gp3csv <TAESC.gp3>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, v := range args {
		err := processArchive(v, *tadaOpts, *unitsBool, *weaponsBool, *imgBool, *tadaBool)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
	}
}

func processArchive(path, topts string, u, w, p, t bool) error {
	var (
		modUnitsDir     = ""
		modWeaponDir    = ""
		modUnitpicDir   = ""
		modDownloadsDir = ""
	)
	a, err := os.Open(path)
	if err != nil {
		return err
	}
	db, err := hpi.LoadHPI(a)
	if err != nil {
		return err
	}
	for k := range db {
		dir := strings.Split(k, "/")

		if len(dir) > 1 {
			if i := strings.Index(strings.ToLower(dir[1]), unitsDir); i == 0 {
				modUnitsDir = dir[1]
				continue
			}
			if i := strings.Index(strings.ToLower(dir[1]), weaponDir); i == 0 {
				modWeaponDir = dir[1]
				continue
			}
			if i := strings.Index(strings.ToLower(dir[1]), unitpicDir); i == 0 {
				modUnitpicDir = dir[1]
				continue
			}
			if i := strings.Index(strings.ToLower(dir[1]), downloadsDir); i == 0 {
				modDownloadsDir = dir[1]
				continue
			}
			if modUnitsDir != "" && modWeaponDir != "" && modUnitpicDir != "" && modDownloadsDir != "" {
				break
			}
		}
	}
	if modUnitsDir == "" {
		fmt.Fprint(os.Stderr, "could not find units directory")
		os.Exit(1)
	}
	if modWeaponDir == "" {
		fmt.Fprint(os.Stderr, "could not find weapon directory")
		os.Exit(1)
	}
	if modUnitpicDir == "" {
		fmt.Fprint(os.Stderr, "could not find unitpic directory")
		os.Exit(1)
	}
	if modDownloadsDir == "" && !t {
		fmt.Fprint(os.Stderr, "could not find downloads directory")
		os.Exit(1)
	}
	ugd.SetPaths(modUnitsDir, modWeaponDir, modUnitpicDir, modDownloadsDir)
	if t && topts != "" {
		opts := strings.Split(topts, ",")
		if len(opts) != 4 {
			fmt.Fprint(os.Stderr, "invalid topts argument\n")
			os.Exit(1)
		}
		cassetteZip, err := os.Create(opts[1] + ".zip")
		if err != nil {
			return err
		}
		err = ugd.CreateModCassette(db, ugd.Manifest{
			ModName:   opts[0],
			Checksum:  opts[1],
			Extension: opts[2],
			Features:  strings.Split(opts[3], " "),
		}, cassetteZip)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		os.Exit(0)
	}
	if u {
		unitCSV, err := os.Create(path + "_units.csv")
		if err != nil {
			return err
		}
		err = ugd.EncodeUnitsCSV(db, unitCSV)
		if err != nil {
			return err
		}
		unitCSV.Close()
	}
	if w {
		weapCSV, err := os.Create(path + "_weap.csv")
		if err != nil {
			return err
		}
		err = ugd.EncodeWeaponsCSV(db, weapCSV)
		if err != nil {
			return err
		}
		weapCSV.Close()
	}
	if p {
		picZip, err := os.Create(path + "_pics.zip")
		if err != nil {
			return err
		}
		pics, err := ugd.GatherUnitPics(db)
		if err != nil {
			return err
		}
		err = ugd.ExportPicsToZip(pics, picZip)
		if err != nil {
			return err
		}
		picZip.Close()
	}
	return err
}
