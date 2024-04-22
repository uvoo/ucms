package config

import (
	"os"
	"path/filepath"
)

// var HomeDir string
/*
   HomeDir, err := os.UserHomeDir()
   if err != nil {
       fmt.Println("Error:", err)
       return
   }
*/
// var HomeDir, err := os.UserHomeDir()
var HomeDir, _ = os.UserHomeDir()
var AppDir string = filepath.Join(HomeDir, ".ucms")

var GeoLite2CountryMMDBFile string = filepath.Join(AppDir, "GeoLite2-Country.mmdb")
