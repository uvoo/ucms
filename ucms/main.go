package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	// "io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// _ "github.com/mattn/go-sqlite3"
	// "os"

	"github.com/gomarkdown/markdown"
	// "github.com/gomarkdown/markdown/ast"
	"bytes"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4/middleware"
	"github.com/microcosm-cc/bluemonday"
	html_template "html/template"
	"regexp"
	"strconv"
	"text/template"
	"uvoo.io/ucms/html_templates"

	"errors"
	"flag"
	// "fmt"
	"log"
	// "net/http"
	"sync"

	// "github.com/labstack/echo/v4"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pquerna/otp/totp"
	// "net/url"
	"uvoo.io/ucms/internal/models"
	"uvoo.io/ucms/internal/database"
)

type NetIPNet struct {
        *net.IPNet
}

var recaptchav3SiteKey string

const uploadUUID = "b28d974e-f742-11ee-950e-63fcdb6c8fb4"

// var db *gorm.DB

// Contains checks if the IP is within the subnet.
func (n NetIPNet) Contains(ip net.IP) bool {
	return n.IPNet.Contains(ip)
}

func WIPGetIPCityISOCode() {
	// Open the MaxMind GeoIP2 City database
	db, err := maxminddb.Open("GeoIP2-City.mmdb")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// IP address to look up
	ip := net.ParseIP("8.8.8.8") // Example IP address, you can change it to any IP you want to look up

	// Define a struct to store the result of the lookup
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
		return
	}

	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	fmt.Printf("IP address %s is located in %s, %s\n", ip, cityName, countryName)
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

func GetIPCountryISOCode(ipString string) (string, error) {
	ipAddr := net.ParseIP(ipString)
	if ipAddr == nil {
		msg := fmt.Sprintf("Invalid IP address:", ipAddr)
		return "", errors.New(msg)
	}

	if isPrivateIP(ipAddr) == true {
		return "Private", nil
	}

	db, err := maxminddb.Open("GeoLite2-Country.mmdb")
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

func isUUID(str string) bool {
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	regex := regexp.MustCompile(uuidPattern)
	return regex.MatchString(str)
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func authenticate(username, password string, c echo.Context) (bool, error) {
	if username == "username" && password == "password" {
		return true, nil
	}
	return false, nil
}

func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func getHTML(body string, title string, tplName string) (string, error) {

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

func getSubnetFromIP(ipString string) (*net.IPNet, error) {
	// func getSubnetFromIP(ipString string) (net.IP, error) {
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
	// Parse subnets
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

func allowIP(clientIP string) bool {
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

	return false
}

func getPage(c echo.Context) error {
	path := c.Request().URL.Path
	fmt.Println("path: ", path)
	id := c.Param("id")
	var page models.Page
	if isUUID(id) {
		if err := database.DBCon.Where("id = ?", id).First(&page).Error; err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
	} else {
		if err := database.DBCon.Where("name = ?", id).First(&page).Error; err != nil {
			return c.String(http.StatusNotFound, "Page not found")
		}
	}

	var body string
	if page.Template == "markdown" {
		md := []byte(page.Content)
		maybeUnsafeHTML := markdown.ToHTML(md, nil, nil)
		tmp := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
		body = fmt.Sprintf("%s", tmp)
	} else {
		body = fmt.Sprintf("%s", page.Content)
	}
	html, err := getHTML(fmt.Sprintf("%s", body), "test1", fmt.Sprintf("%s", page.Template))
	if err != nil {
		fmt.Println("err %s", err)
	}

	return c.HTML(http.StatusOK, fmt.Sprintf("%s", html))
}

func updatePage(c echo.Context) error {
	fmt.Println("pagefoo")
	id := c.Param("id")
	page := new(models.Page)

	if err := database.DBCon.First(&page, "id = ?", id).Error; err != nil {
		return err
	}

	if err := c.Bind(page); err != nil {
		return err
	}

	if err := database.DBCon.Save(page).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, page)
}

func createPage(c echo.Context) error {
	page := new(models.Page)
	if err := c.Bind(page); err != nil {
		return err
	}
	if page.ID == uuid.Nil {
		page.ID = uuid.New()
	}
	database.DBCon.Create(&page)
	return c.JSON(http.StatusCreated, page)
}

func uploadFile(c echo.Context) error {
	fmt.Println("uploadfoo")
	// Read form data
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination file
	dst, err := os.Create(filepath.Join("assets", file.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the file to the destination
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.String(http.StatusOK, "File uploaded successfully")
}

// Function to handle file download
func downloadFile(c echo.Context) error {
	filename := c.Param("file")
	filePath := filepath.Join("assets", filename)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Stream the file to the client
	return c.File(filePath)
}

func xgetEnvOrDefault(envVarName string, defaultValue string) string {
	envVarValue, exists := os.LookupEnv(envVarName)

	if exists {
		return envVarValue
	} else {
		return defaultValue
	}
}

func getEnvOrDefault(envVarName string, defaultValue string, required bool) (string, error) {
	envVarValue, exists := os.LookupEnv(envVarName)
	if exists {
		return envVarValue, nil
	}

	if !required {
		return defaultValue, nil
	}

	// if defaultValue == "" || defaultValue == nil {
	if defaultValue == "" {
		return "", errors.New("environment variable not set and no default value provided")
	}

	return defaultValue, nil
}

// func getSubmitPage(){
// getSumitPage(recaptchav3SiteKey string){
func getSubmitPage(c echo.Context) error {
	// Example key-value pairs
	data := map[string]interface{}{
		"Recaptchav3SiteKey": recaptchav3SiteKey,
	}

	// Render the template file "template.tmpl" with the given data
	// renderedTemplate, err := RenderTemplate("template.tmpl", data)
	// renderedTemplate, err := RenderTemplate(html_templates.Submit, data)
	renderedTemplate, err := RenderTemplate("templates/submit.html", data)
	if err != nil {
		log.Fatalf("Error rendering template: %v", err)
	}

	// Print the rendered template
	// fmt.Println(renderedTemplate)
	// return renderedTemplate
	return c.HTML(http.StatusOK, fmt.Sprintf("%s", renderedTemplate))
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

func filterIP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		clientIP := c.RealIP()
		// fmt.Println("ip:", clientIP)

		if allowIP(clientIP) {
			return next(c)
		}
		// return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		return echo.NewHTTPError(http.StatusUnauthorized,
			fmt.Sprintf("IP address %s not allowed", c.RealIP()))
	}
}

func startServer(port string, isTLS bool, certFile, keyFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.IPExtractor = middleware.RealIPWithUntrustedHeader
	e.Use(filterIP)

	r := e.Group("/secure")
	r.Use(middleware.JWT([]byte("secret")))
	r.GET("", func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		return c.String(http.StatusOK, "Welcome "+username+"!")
	})

	authRoutes := e.Group("")
	authRoutes.Use(middleware.BasicAuth(authenticate))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("ok"))
	})

	authRoutes.GET("/download/:file", downloadFile)

	e.GET("/editor", func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf("%s", html_templates.Editor))
	})

	e.GET("/fwtest", func(c echo.Context) error {
		clientIPAddress := c.RealIP()
		countryCode, err := GetIPCountryISOCode(clientIPAddress)
		// fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode)
		if err != nil {
			fmt.Println(err)
		}

		return c.String(http.StatusOK, fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode))
	})

	e.GET("/ip", func(c echo.Context) error {
		clientIPAddress := c.RealIP()
		countryCode, err := GetIPCountryISOCode(clientIPAddress)
		if err != nil {
			fmt.Println(err)
		}
		return c.String(http.StatusOK, fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode))
	})

	e.POST("/login", postLogin)
	authRoutes.PATCH("/page/:id", updatePage)
	authRoutes.POST("/page", createPage)
	e.GET("/page/:id", getPage)

	authRoutes.POST("/upload", uploadFile)

	// e.GET("/submit", getSumitPage(recaptchav3SiteKey))
	e.GET("/submit", getSubmitPage)
	//e.GET("/submit", func(c echo.Context) error {
	//		return c.HTML(http.StatusOK, fmt.Sprintf("%s", html_templates.Submit))
	//	})

	e.POST("/submit", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		user := models.User{Name: name, Email: email}

		if err := database.DBCon.Create(&user).Error; err != nil {
			return err
		}

		return c.String(http.StatusOK, "User successfully created")
	})

	e.GET("/x-forwarded-port", func(c echo.Context) error {
		xForwardedPort := c.Request().Header.Get("X-Forwarded-Port")
		return c.String(http.StatusOK, fmt.Sprintf("X-Forwarded-Port: %s", xForwardedPort))
	})

	// Bind Listeners
	address := fmt.Sprintf(":%s", port)
	fmt.Printf("Server is listening on port %s\n", port)
	if isTLS {
		if err := e.StartTLS(address, certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTPS server on port %s: %v", port, err)
		}
	} else {
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server on port %s: %v", port, err)
		}
	}

}

func postLogin(c echo.Context) error {
	fmt.Println("foo")
	// e.POST("/login", func(c echo.Context) error {
	// Get username and password from request
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Authenticate user (this is a simple example, you should hash and compare passwords securely)
	var user models.User
	if err := database.DBCon.Where("username = ?", username).First(&user).Error; err != nil {
		fmt.Println(err)
		return echo.ErrUnauthorized
	}
	if user.Password != password {
		fmt.Println("unauth")
		return echo.ErrUnauthorized
	}

	// Generate OTP secret and store it
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "YourAppName",
		AccountName: user.Username,
	})
	if err != nil {
		return err
	}
	user.Secret = secret.Secret()
	database.DBCon.Save(&user)

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JwtCustomClaims{
		ID:       user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(), // Token valid for 3 days
		},
	})

	// Sign the token with your secret
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	// Return the token
	return c.JSON(http.StatusOK, map[string]string{
		"token": tokenString,
	})
}

func main() {
	httpPort := flag.String("http-port", "18080", "HTTP port number")
	httpsPort := flag.String("https-port", "18443", "HTTPS port number")
	flag.Parse()

	certFile := "cert.pem"
	keyFile := "key.pem"

	recaptchav3SiteKey, _ = getEnvOrDefault("RECAPTCHAV3_SITE_KEY", "", true)

	var err error
	database.DBCon, err = gorm.Open(sqlite.Open("ucms.db"), &gorm.Config{})
	if err != nil {
		panic(err)
		// log.Fatal(err)
	}
	models.Migrate()

	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(*httpPort, false, "", "", &wg)

	wg.Add(1)
	go startServer(*httpsPort, true, certFile, keyFile, &wg)

	wg.Wait()
}
