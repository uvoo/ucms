package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	// "os"

	"github.com/gomarkdown/markdown"
	// "github.com/gomarkdown/markdown/ast"
	"bytes"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4/middleware"
	"github.com/microcosm-cc/bluemonday"
	"text/template"
)

type Page struct {
	ID       int    `gorm:"primary_key"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Template string `json:"template"`
}

/*
    HTMLType id
type HTMLType struct {
  ID   int
  Name string
}
*/

func authenticate(username, password string, c echo.Context) (bool, error) {
	// Check username and password here, e.g., from a database or predefined credentials
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
	const tplMarkdown = `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{{.Title}}</title>
    </head>
    <body>
	   {{ .Body }}
    </body>
    </html>
`
	const tplMDBootstrap = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Hello World</title>
		<!-- Include Bootstrap 5 CSS -->
		<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/css/bootstrap.min.css" rel="stylesheet">
		<!-- Include MDBootstrap CSS -->
		<link href="https://cdnjs.cloudflare.com/ajax/libs/mdbootstrap/4.19.1/css/mdb.min.css" rel="stylesheet">
		<!-- Include FontAwesome CSS -->
		<link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css" rel="stylesheet">
	</head>
    <body>
	   {{ .Body }}
	   <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/js/bootstrap.bundle.min.js"></script>
    </body>
	</html>
  `
	const tplError = `
  E: Unsupported template
  `
/*
	<body>
		<div class="container mt-5">
			<h1>Hello, World!</h1>
			<p>This is a simple HTML page served by Labstack Echo with Bootstrap 5, MDBootstrap, and FontAwesome.</p>
			<i class="fas fa-smile"></i> <!-- FontAwesome icon -->
		</div>

		<!-- Include Bootstrap 5 JS (Optional) -->
		<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/js/bootstrap.bundle.min.js"></script>
	</body>
*/
	data := struct {
		Title string
		Body  string
	}{
		Title: title,
		Body:  body,
	}

	var tplContent string

	// Determine the template content based on the template name
	switch tplName {
	case "markdown":
		tplContent = tplMarkdown
	case "mdbootstrap":
		tplContent = tplMDBootstrap
	default:
		tplContent = tplError
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

func main() {
	db, err := gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.AutoMigrate(&Page{})

	e := echo.New()
	protectedRoutes := e.Group("")
	protectedRoutes.Use(middleware.BasicAuth(authenticate))

	// e.Use(middleware.BasicAuth(authenticate))


    e.GET("/editor", func(c echo.Context) error {
    const editorHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WYSIWYG Editor Example</title>
    <!-- Include CKEditor -->
    <script src="https://cdn.ckeditor.com/ckeditor5/37.0.0/classic/ckeditor.js"></script>
</head>                                                                                                                                                                                      <body>                                                                                                                                                                                           <h1>WYSIWYG Editor Example</h1>                                                                                                                                                              <form action="/submit" method="POST">
        <textarea id="editor" name="content"></textarea>
        <button type="submit">Submit</button>
    </form>                                                                                                                                                                                                                                                                                                                                                                                   <script>                                                                                                                                                                                         // Initialize CKEditor with source editing enabled
        ClassicEditor
            .create(document.querySelector('#editor'), {
                toolbar: ['', 'heading', '|', 'bold', 'italic', 'link', 'bulletedList', 'numberedList', '|', 'indent', 'outdent', '|', 'blockQuote', 'insertTable', '|', 'undo', 'redo', '|', 'source', 'uploadImage', 'blockQuote', 'codeBlock'],
/*
toolbar: {
    items: [
        'undo', 'redo',
        '|', 'heading',
        '|', 'fontfamily', 'fontsize', 'fontColor', 'fontBackgroundColor',
        '|', 'bold', 'italic', 'strikethrough', 'subscript', 'superscript', 'code',
        '|', 'link', 'uploadImage', 'blockQuote', 'codeBlock',
        '|', 'bulletedList', 'numberedList', 'todoList', 'outdent', 'indent'
    ],
    shouldNotGroupWhenFull: false
}
*/
                language: 'en'
            })
            .catch(error => {
                console.error(error);
            });
    </script>                                                                                                                                                                                </body>
</html>
    `

        return c.HTML(http.StatusOK, fmt.Sprintf("%s", editorHTML))
    })


	e.GET("/page/:id", func(c echo.Context) error {
		id := c.Param("id")
		var page Page
		if err := db.First(&page, id).Error; err != nil {
			return err
		}
		var body string
		if page.Template == "markdown" {
		md := []byte(page.Content)
		maybeUnsafeHTML := markdown.ToHTML(md, nil, nil)
		tmp := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
		body = fmt.Sprintf("%s", tmp)
		} else {
		// body := bluemonday.UGCPolicy().SanitizeBytes(page.Content)
		// body = template.HTMLEscapeString(fmt.Sprintf("%s", page.Content))
		body = fmt.Sprintf("%s", page.Content)
		}
		// html, err := getHTML(fmt.Sprintf("%s", body), "test1", "markdown")
		// html, err := getHTML(fmt.Sprintf("%s", body), "test1", "mdbootstrap")
		html, err := getHTML(fmt.Sprintf("%s", body), "test1", fmt.Sprintf("%s", page.Template))
		if err != nil {
			fmt.Println("err %s", err)
		}

		return c.HTML(http.StatusOK, fmt.Sprintf("%s", html))
	})
	// }, middleware.BasicAuth(authentication))

	protectedRoutes.POST("/page", func(c echo.Context) error {
		markdown := new(Page)
		if err := c.Bind(markdown); err != nil {
			return err
		}
		db.Create(&markdown)
		return c.JSON(http.StatusCreated, markdown)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
