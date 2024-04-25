package handlers

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"uvoo.io/ucms/internal/database"
	"uvoo.io/ucms/internal/models"
	"uvoo.io/ucms/internal/utils"
	"uvoo.io/ucms/internal/templates"
	// "github.com/gomarkdown/markdown/html"
	"github.com/golang-jwt/jwt"
	"github.com/pquerna/otp/totp"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// var recaptchav3SiteKey string

/*
   // "io/ioutil"
   "net/http"

   "github.com/labstack/echo/v4"
   "gorm.io/driver/sqlite"
   "gorm.io/gorm"
   // _ "github.com/mattn/go-sqlite3"
   // "os"

   // "github.com/gomarkdown/markdown/ast"
   "bytes"
   "github.com/gomarkdown/markdown/parser"
   "github.com/labstack/echo/v4/middleware"
   html_template "html/template"
   "regexp"
   "strconv"
   "text/template"
   "uvoo.io/ucms/html_templates"

   "errors"
   "flag"
   // "fmt"
   // "net/http"
   "sync"

   // "github.com/labstack/echo/v4"
   "github.com/oschwald/maxminddb-golang"
   "net"
   "strings"
   "time"

*/

func GetPage(c echo.Context) error {
	path := c.Request().URL.Path
	fmt.Println("path: ", path)
	id := c.Param("id")
	var page models.Page
	if utils.IsUUID(id) {
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
	html, err := utils.GetHTML(fmt.Sprintf("%s", body), "test1", fmt.Sprintf("%s", page.Template))
	if err != nil {
		fmt.Println("err %s", err)
	}

	return c.HTML(http.StatusOK, fmt.Sprintf("%s", html))
}

func UpdatePage(c echo.Context) error {
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

func CreatePage(c echo.Context) error {
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

func UploadFile(c echo.Context) error {
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
func DownloadFile(c echo.Context) error {
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

// func GetSubmit(c echo.Context) error {
func GetSubmit(c echo.Context, recaptchav3SiteKey string) error {
	// recaptchav3SiteKey, _ = utils.GetEnvOrDefault("RECAPTCHAV3_SITE_KEY", "", true)
	data := map[string]interface{}{
		"Recaptchav3SiteKey": recaptchav3SiteKey,
		// "Recaptchav3SiteKey": GetRecaptchav3SiteKey(),
	}
	// "Recaptchav3SiteKey": recaptchav3SiteKey,

	// Render the template file "template.tmpl" with the given data
	// renderedTemplate, err := RenderTemplate("template.tmpl", data)
	// renderedTemplate, err := RenderTemplate(html_templates.Submit, data)
	renderedTemplate, err := utils.RenderTemplate("templates/submit.html", data)
	if err != nil {
		log.Fatalf("Error rendering template: %v", err)
	}

	// Print the rendered template
	// fmt.Println(renderedTemplate)
	// return renderedTemplate
	return c.HTML(http.StatusOK, fmt.Sprintf("%s", renderedTemplate))
}

func PostLogin(c echo.Context) error {
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
