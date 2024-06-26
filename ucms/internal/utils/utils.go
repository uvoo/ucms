package utils

import (
	"bytes"
	"errors"
	"fmt"
	html_template "html/template"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
	"github.com/oschwald/maxminddb-golang"

	"uvoo.io/ucms/html_templates"
	"uvoo.io/ucms/internal/config"
	"uvoo.io/ucms/internal/database"
	"uvoo.io/ucms/internal/models"
)


type IPInfo struct {
    CountryISOCode string
    Subdivision    string
    City           string
}


func isPrivateIP(ip net.IP) bool {
	privateIPv4Blocks := []*net.IPNet{
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
	}

	privateIPv6Blocks := []*net.IPNet{
		{IP: net.ParseIP("::1"), Mask: net.CIDRMask(7, 128)},
		{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)},
		{IP: net.ParseIP("fe80::"), Mask: net.CIDRMask(10, 128)},
	}

	for _, block := range privateIPv4Blocks {
		if block.Contains(ip) {
			return true
		}
	}

	for _, block := range privateIPv6Blocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}

func GetHTML(body string, title string, tplName string) (string, error) {

	data := struct {
		Title string
		Body  string
	}{
		Title: title,
		Body:  body,
	}

	var tplContent string

	switch tplName {
	case "markdown":
		tplContent = html_templates.Markdown
	case "mdbootstrap":
		tplContent = html_templates.MDBootstrap
	default:
		tplContent = html_templates.Error
	}

	t := template.Must(template.New(tplName).Parse(tplContent))

	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	output := buf.String()
	return output, nil
}

func GetSubnetFromIP(ipString string) (*net.IPNet, error) {
	if !strings.Contains(ipString, "/") {
		if strings.Contains(ipString, ":") {
			ipString += "/128" // IPv6 address
		} else {
			ipString += "/32" // IPv4 address
		}
	}
	_, subnet, err := net.ParseCIDR(ipString)
	if err != nil {
		fmt.Println("Invalid CIDR notation:", err)
		return nil, err
	}

	return subnet, nil
}

func subnetContains(subnet1, subnet2 string) bool {
	_, subnetObj1, err := net.ParseCIDR(subnet1)
	if err != nil {
		return false
	}

	_, subnetObj2, err := net.ParseCIDR(subnet2)
	if err != nil {
		return false
	}

	return subnetObj1.Contains(subnetObj2.IP)
}

func toIPNet(cidr string) net.IPNet {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return *ipnet
}

func CheckIPInSubnet(ipStr string, subnetStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false, fmt.Errorf("invalid subnet: %s", subnetStr)
	}

	return subnet.Contains(ip), nil
}

func AllowIP(clientIP string) bool {
	var fwRules []models.FWRule
	database.DBCon.Order("priority ASC").Find(&fwRules)
	// net.ParseIP(clientIPAddress)
	for _, fwRule := range fwRules {
		// fmt.Println("fwRule: %v", fwRule)
		var srcIPInSubnet bool
		srcIPInSubnet, _ = CheckIPInSubnet(clientIP, fwRule.SrcIPNet)
		if fwRule.Active == true && fwRule.Action == models.Deny && srcIPInSubnet == true {
			return false
		} else if fwRule.Active == true && fwRule.Action == models.Allow && srcIPInSubnet == true {
			return true
		} else {
			return false
		}
	}

	clientCountryCode, err := GetIPCountryISOCode(clientIP)
	if err != nil {
		msg := fmt.Sprintf("Error:", err)
		fmt.Println(msg)
	}
	var countryCodeRules []models.CountryCodeRule
	database.DBCon.Order("priority ASC").Find(&countryCodeRules)
	for _, countryCodeRule := range countryCodeRules {
		if countryCodeRule.Code == clientCountryCode {
			return true
		}
	}

	clientCityCode, clientCountryCode, err := GetIPCityCountryISOCode(clientIP)
	if err != nil {
		msg := fmt.Sprintf("Error:", err)
		fmt.Println(msg)
	}
	var cityCodeRules []models.CityCodeRule
	database.DBCon.Order("priority ASC").Find(&cityCodeRules)
	for _, cityCodeRule := range cityCodeRules {
		if cityCodeRule.Code == clientCityCode {
			return true
		}
	}

	return false
}

func IsUUID(str string) bool {
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	regex := regexp.MustCompile(uuidPattern)
	return regex.MatchString(str)
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func GetIPCountryISOCode(ipString string) (string, error) {
	ipAddr := net.ParseIP(ipString)
	if ipAddr == nil {
		msg := fmt.Sprintf("Invalid IP address:", ipAddr)
		return "", errors.New(msg)
	}

	if isPrivateIP(ipAddr) == true {
		return "Private", nil
	}

	db, err := maxminddb.Open(config.GeoLite2CountryMMDBFile)
	if err != nil {
		msg := fmt.Sprintf("Error opening database:", err)
		return "", errors.New(msg)
	}
	defer db.Close()

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	err = db.Lookup(ipAddr, &record)
	if err != nil {
		msg := fmt.Sprintf("Error looking up IP address:", err)
		return "", errors.New(msg)
	}

	return record.Country.ISOCode, nil
}

func GetIPCityCountryISOCode(ipString string) (string, string, error) {
	db, err := maxminddb.Open(config.GeoLite2CityMMDBFile)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return "", "", err
	}
	defer db.Close()

	ip := net.ParseIP(ipString)

	var record struct {
		City struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"city"`
		Country struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"country"`
	}

	err = db.Lookup(ip, &record)
	if err != nil {
		fmt.Println("Error looking up IP address:", err)
		return "", "", err
	}

	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	fmt.Printf("IP address %s is located in %s, %s\n", ip, cityName, countryName)
	return cityName, countryName, nil
}

/*
func GetIPGeoInfo(ipString string) (string, string, string, error) {
	db, err := maxminddb.Open(config.GeoLite2CityMMDBFile)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return "", "", "", err
	}
	defer db.Close()

	ip := net.ParseIP(ipString)

	var record struct {
		City struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"city"`
		Country struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"country"`
	}

	err = db.Lookup(ip, &record)
	if err != nil {
		fmt.Println("Error looking up IP address:", err)
		return "", "", "", err
	}

	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	stateName := record.Subdivisions.Names["en"]
	countryISOCode := record.Country.iso_code
	fmt.Printf("IP address %s is located in %s, %s\n", ip, cityName, countryName)
    return countryISOCode, stateName, cityName, nil

/*
    e.GET("/", func(c echo.Context) error {
        // Create a map to hold the JSON response
        response := map[string]string{
            "country_iso_code": countryISOCode,
            "state": stateName,
            "city": cityName,
        }
        // Return JSON response with status code 200
        return c.JSON(http.StatusOK, response)
    })
*/
// }

func GetIPGeoInfo(ipAddress string) (*IPInfo, error) {
    // Open the MaxMind GeoLite2-City database
    db, err := maxminddb.Open(config.GeoLite2CityMMDBFile)
    if err != nil {
        return nil, err
    }
    defer db.Close()

    // Parse the IP address
    ip := net.ParseIP(ipAddress)
    if ip == nil {
        return nil, errors.New("invalid IP address")
    }

    // Lookup the IP address in the database
    var record struct {
        Country struct {
            ISOCode string `maxminddb:"iso_code"`
        } `maxminddb:"country"`
        Subdivisions []struct {
            Name string `maxminddb:"name"`
        } `maxminddb:"subdivisions"`
        City struct {
            Name string `maxminddb:"name"`
        } `maxminddb:"city"`
    }

    if err := db.Lookup(ip, &record); err != nil {
        return nil, err
    }

    // Create and populate the IPInfo struct
    ipInfo := IPInfo{
        CountryISOCode: record.Country.ISOCode,
        City:           record.City.Name,
    }

    // Concatenate subdivision names if available
    if len(record.Subdivisions) > 0 {
        for _, sub := range record.Subdivisions {
            ipInfo.Subdivision += sub.Name + ", "
        }
        // Remove the trailing comma and space
        ipInfo.Subdivision = ipInfo.Subdivision[:len(ipInfo.Subdivision)-2]
    }

    return &ipInfo, nil
}

/*
func GetIPGeoInfo(ipAddress string) (*IPInfo, error) {
type IPInfo struct {
    CountryISOCode string
    Subdivision    string
    City           string
}
    // Open the MaxMind GeoLite2-City database
    db, err := maxminddb.Open("GeoLite2-City.mmdb")
    if err != nil {
        return nil, err
    }
    defer db.Close()

    // Parse the IP address
    ip := net.ParseIP(ipAddress)
    if ip == nil {
        return nil, errors.New("invalid IP address")
    }

    // Lookup the IP address in the database
    var record struct {
        Country struct {
            ISOCode string `maxminddb:"iso_code"`
        } `maxminddb:"country"`
        Subdivisions []struct {
            Name string `maxminddb:"name"`
        } `maxminddb:"subdivisions"`
        City struct {
            Name string `maxminddb:"name"`
        } `maxminddb:"city"`
    }

    if err := db.Lookup(ip, &record); err != nil {
        return nil, err
    }

    // Create and populate the IPInfo struct
    ipInfo := IPInfo{
        CountryISOCode: record.Country.ISOCode,
        City:           record.City.Name,
    }

    // Concatenate subdivision names if available
    if len(record.Subdivisions) > 0 {
        for _, sub := range record.Subdivisions {
            ipInfo.Subdivision += sub.Name + ", "
        }
        // Remove the trailing comma and space
        ipInfo.Subdivision = ipInfo.Subdivision[:len(ipInfo.Subdivision)-2]
    }

    return &ipInfo, nil
}
*/

func Authenticate(username, password string, c echo.Context) (bool, error) {
	if username == "username" && password == "password" {
		return true, nil
	}
	return false, nil
}

func MDToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func GetEnvOrDefault(envVarName string, defaultValue string, required bool) (string, error) {
	envVarValue, exists := os.LookupEnv(envVarName)
	if exists {
		return envVarValue, nil
	}

	if !required {
		return defaultValue, nil
	}

	if defaultValue == "" {
		// return "", errors.New("environment variable %s not set and no default value provided", envVarName)
		return "", errors.New(fmt.Sprintf("Environment variable %s is not set.", envVarName))
		// Log.error("E: Environment variable %s is not set.", envVarName))
		// panic(err)
	}

	return defaultValue, nil
}

func RenderTemplate(templateFile string, data map[string]interface{}) (string, error) {
	tmpl, err := html_template.ParseFiles(templateFile)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func FilterIP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientIP := c.RealIP()

		if AllowIP(clientIP) {
			return next(c)
		}
		// return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		return echo.NewHTTPError(http.StatusUnauthorized,
			fmt.Sprintf("IP address %s not allowed", c.RealIP()))
	}
}
