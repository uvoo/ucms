package main

import (
    "net/http"
    "text/template"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Define HTML template
    tpl := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{{.Title}}</title>
    </head>
    <body>
        <h1>Hello, World!</h1>
        <button onclick="changeTitle()">Change Title</button>
        <script>
            function changeTitle() {
                document.title = "New Title";
            }
        </script>
    </body>
    </html>
    `

    // Execute the template
    t := template.Must(template.New("html").Parse(tpl))
    data := struct{ Title string }{Title: "Original Title"}
    t.Execute(w, data)
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

