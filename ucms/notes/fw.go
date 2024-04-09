package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/oschwald/maxminddb-golang"
	"net"
)

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

	// Perform the lookup
	err = db.Lookup(ip, &record)
	if err != nil {
		fmt.Println("Error looking up IP address:", err)
		return
	}

	// Print the result
	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	fmt.Printf("IP address %s is located in %s, %s\n", ip, cityName, countryName)
}

func isPrivateIP(ip net.IP) bool {
	// fmt.Println("IPAddress: %v", ip)
	privateIPv4Blocks := []*net.IPNet{
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
	}

	// Private IPv6 addresses (ULA - Unique Local Addresses and Link-local)
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

func GetIPCountryISOCode(ipAddress string) (string, error) {
	ipo := net.ParseIP(ipAddress)
	if isPrivateIP(ipo) == true {
		return "Private", nil
	}
	db, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		msg := fmt.Sprintf("Error opening database:", err)
		return "", errors.New(msg)
	}
	defer db.Close()

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		msg := fmt.Sprintf("Invalid IP address:", ipAddress)
		return "", errors.New(msg)
	}

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	err = db.Lookup(ip, &record)
	if err != nil {
		msg := fmt.Sprintf("Error looking up IP address:", err)
		return "", errors.New(msg)
	}
	return record.Country.ISOCode, nil
}

func startServer(port string, isTLS bool, certFile, keyFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create a new Echo instance
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			notAllowedIPAddress := "192.168.1.100" // Set the allowed IP address here

			// clientIP := c.RealIP()

			clientIPAddress := c.RealIP()
			countryCode, err := GetIPCountryISOCode(clientIPAddress)
			if err != nil {
				msg := fmt.Sprintf("Error:", err)
				fmt.Println(msg)
			}
			if clientIPAddress == notAllowedIPAddress || countryCode != "US" {
				return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
			}

			return next(c)
		}
	})

	// Define your routes
	e.GET("/", func(c echo.Context) error {
		xForwardedPort := c.Request().Header.Get("X-Forwarded-Port")
		// return c.String(http.StatusOK, fmt.Sprintf("port: %s", port))
		return c.String(http.StatusOK, fmt.Sprintf("port: %s", xForwardedPort))
	})

	e.GET("/x-forwarded-port", func(c echo.Context) error {
		// Retrieve X-Forwarded-Port header from the request
		xForwardedPort := c.Request().Header.Get("X-Forwarded-Port")

		// Return the X-Forwarded-Port as response
		return c.String(http.StatusOK, fmt.Sprintf("X-Forwarded-Port: %s", xForwardedPort))
	})

	e.GET("/ip", func(c echo.Context) error {
		// Retrieve client's IP address
		clientIPAddress := c.RealIP()
		countryCode, err := GetIPCountryISOCode(clientIPAddress)
		if err != nil {
			fmt.Println(err)
		}

		// Return the IP address as response
		return c.String(http.StatusOK, fmt.Sprintf("IP: %s Country: %s", clientIPAddress, countryCode))
	})

	// Start the server
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
	// Define flags for HTTP and HTTPS ports
	httpPort := flag.String("http-port", "8080", "HTTP port number")
	httpsPort := flag.String("https-port", "8443", "HTTPS port number")
	flag.Parse()

	// Define paths to certificate and key files
	certFile := "cert.pem"
	keyFile := "key.pem"

	// Create a WaitGroup to wait for all servers to start
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go startServer(*httpPort, false, "", "", &wg)

	// Start HTTPS server
	wg.Add(1)
	go startServer(*httpsPort, true, certFile, keyFile, &wg)

	// Wait for all servers to start
	wg.Wait()
}
