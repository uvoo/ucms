package main

import (
	"fmt"
	// "io"
	// "os"
	// "path/filepath"
	// "io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// _ "github.com/mattn/go-sqlite3"
	// "os"

	// "github.com/gomarkdown/markdown"
	// "github.com/gomarkdown/markdown/ast"
	// "bytes"
	// "github.com/gomarkdown/markdown/html"
	// "github.com/gomarkdown/markdown/parser"
	// "github.com/google/uuid"
	"github.com/labstack/echo/v4/middleware"
	// "github.com/microcosm-cc/bluemonday"
	// html_template "html/template"
	// "regexp"
	// "strconv"
	// "text/template"
	"uvoo.io/ucms/html_templates"

	// "errors"
	"flag"
	// "fmt"
	"log"
	// "net/http"
	"sync"

	// "github.com/labstack/echo/v4"
	"github.com/oschwald/maxminddb-golang"
	"net"
	// "strings"
	// "time"

	"github.com/golang-jwt/jwt"
	// "github.com/pquerna/otp/totp"
	// "net/url"
	"uvoo.io/ucms/internal/models"
	"uvoo.io/ucms/internal/database"
	"uvoo.io/ucms/internal/handlers"
	"uvoo.io/ucms/internal/utils"
)

// recaptchav3SiteKey, _ = utils.getEnvOrDefault("RECAPTCHAV3_SITE_KEY", "", true)

type NetIPNet struct {
        *net.IPNet
}

// var recaptchav3SiteKey string

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


func startServer(port string, isTLS bool, certFile, keyFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.IPExtractor = middleware.RealIPWithUntrustedHeader
	e.Use(utils.FilterIP)

	r := e.Group("/secure")
	r.Use(middleware.JWT([]byte("secret")))
	r.GET("", func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		username := claims["username"].(string)
		return c.String(http.StatusOK, "Welcome "+username+"!")
	})

	authRoutes := e.Group("")
	authRoutes.Use(middleware.BasicAuth(utils.Authenticate))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("ok"))
	})

	authRoutes.GET("/download/:file", handlers.DownloadFile)

	e.GET("/editor", func(c echo.Context) error {
		return c.HTML(http.StatusOK, fmt.Sprintf("%s", html_templates.Editor))
	})

	e.GET("/fwtest", func(c echo.Context) error {
		clientIPAddress := c.RealIP()
		countryCode, err := utils.GetIPCountryISOCode(clientIPAddress)
		// fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode)
		if err != nil {
			fmt.Println(err)
		}

		return c.String(http.StatusOK, fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode))
	})

	e.GET("/ip", func(c echo.Context) error {
		clientIPAddress := c.RealIP()
		countryCode, err := utils.GetIPCountryISOCode(clientIPAddress)
		if err != nil {
			fmt.Println(err)
		}
		return c.String(http.StatusOK, fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode))
	})

	e.POST("/login", handlers.PostLogin)
	authRoutes.PATCH("/page/:id", handlers.UpdatePage)
	authRoutes.POST("/page", handlers.CreatePage)
	e.GET("/page/:id", handlers.GetPage)

	authRoutes.POST("/upload", handlers.UploadFile)

	// e.GET("/submit", getSumitPage(recaptchav3SiteKey))
	e.GET("/submit", handlers.GetSubmit)
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

func main() {
	httpPort := flag.String("http-port", "18080", "HTTP port number")
	httpsPort := flag.String("https-port", "18443", "HTTPS port number")
	flag.Parse()

	certFile := "cert.pem"
	keyFile := "key.pem"

	// recaptchav3SiteKey, _ = utils.GetEnvOrDefault("RECAPTCHAV3_SITE_KEY", "", true)

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
