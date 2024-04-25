package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oschwald/maxminddb-golang"
	// "github.com/sirupsen/logrus"
	// "github.com/sirupsen/logrus/hooks/syslog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"uvoo.io/ucms/internal/config"
	"uvoo.io/ucms/html_templates"
	"uvoo.io/ucms/internal/database"
	"uvoo.io/ucms/internal/handlers"
	"uvoo.io/ucms/internal/models"
	"uvoo.io/ucms/internal/utils"
)

type NetIPNet struct {
	*net.IPNet
}

var recaptchav3SiteKey string

const uploadUUID = "b28d974e-f742-11ee-950e-63fcdb6c8fb4"

func (n NetIPNet) Contains(ip net.IP) bool {
	return n.IPNet.Contains(ip)
}

func WIPGetIPCityISOCode() {
	db, err := maxminddb.Open("GeoIP2-City.mmdb")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	ip := net.ParseIP("8.8.8.8")

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

	e.GET("/submit", func(c echo.Context) error {
		return handlers.GetSubmit(c, recaptchav3SiteKey)
	})

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
	var err error
	// log.SetFlags(log.LstdFlags | log.LUTC)
	// logrus.SetFormatter(&logrus.TextFormatter{})
	// logrus.SetLevel(logrus.InfoLevel)

	/*
	   	logger := logrus.New()
	   	hook, err := syslog.NewSyslogHook("", "", 0, "")
	   	if err != nil {
	   		logger.Fatal("Failed to initialize syslog hook:", err)
	   	}
	   	logger.AddHook(hook)
	    // logger.Fatal("Error:", err)
	*/

	httpPort := flag.String("http-port", "18080", "HTTP port number")
	httpsPort := flag.String("https-port", "18443", "HTTPS port number")
	flag.Parse()

	certFile := "cert.pem"
	keyFile := "key.pem"

	recaptchav3SiteKey, err = utils.GetEnvOrDefault("RECAPTCHAV3_SITE_KEY", "", true)
	if err != nil {
		log.Fatal("Error:", err)
	}

	database.DBCon, err = gorm.Open(sqlite.Open(config.DBFile), &gorm.Config{})
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
