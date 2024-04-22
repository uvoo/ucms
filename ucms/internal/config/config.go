package config

import (
	"os"
	"path/filepath"
)

var HomeDir, _ = os.UserHomeDir()
var AppDir string = filepath.Join(HomeDir, ".ucms")
var GeoLite2CountryMMDBFile string = filepath.Join(AppDir, "GeoLite2-Country.mmdb")
var GeoLite2CityMMDBFile string = filepath.Join(AppDir, "GeoLite2-City.mmdb")
