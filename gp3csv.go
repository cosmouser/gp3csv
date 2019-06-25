package main

import (
	"github.com/cosmouser/gp3csv/hpi"
	"github.com/cosmouser/gp3csv/ugd"
	"os"
	"fmt"
	"flag"
)

func main() {
	unitsBool := flag.Bool("u", true, "extract unit data")
	weaponsBool := flag.Bool("w", true, "extract weapon data")
	imgBool := flag.Bool("p", true, "extract unitpics")
	versBool := flag.Bool("v", false, "prints the version and exits")
	flag.Parse()
	if *versBool {
		fmt.Println("gp3csv version 1.0.0")
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "gp3csv TAESC.gp3")
		flag.PrintDefaults()
		os.Exit(1)
	}
	for _, v := range args {
		err := processArchive(v, *unitsBool, *weaponsBool, *imgBool)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
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




