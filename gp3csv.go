package main

import (
	"github.com/cosmouser/gp3csv/hpi"
	"github.com/cosmouser/gp3csv/ugd"
	"os"
	"flag"
)

func main() {
	unitsBool := flag.Bool("u", true, "extract unit data")
	weaponsBool := flag.Bool("w", true, "extract weapon data")
	imgBool := flag.Bool("p", true, "extract unitpics")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	for _, v := range args {
		err := processArchive(v, *unitsBool, *weaponsBool, *imgBool)
		if err != nil {
			panic(err)
		}
	}
}

func processArchive(path string, u, w, p bool) error {
	a, err := os.Open(path)
	if err != nil {
		return err
	}
	db, err := hpi.LoadHPI(a)
	if err != nil {
		return err
	}
	if u {
		unitCSV, err := os.Create(path+"_units.csv")
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
		weapCSV, err := os.Create(path+"_weap.csv")
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
		picZip, err := os.Create(path+"_pics.zip")
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




