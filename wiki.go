package main

import (
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) HTMLBody() template.HTML {
	escaped := html.EscapeString(string(p.Body))
	replaced := strings.ReplaceAll(escaped, "\r\n", "<br>")
	replaced = strings.ReplaceAll(replaced, "\n", "<br>")
	return template.HTML(replaced)
}

type PageList struct {
	Pages []string
}

// テンプレートを起動時にキャッシュ
var templates = template.Must(template.ParseFiles("view.html", "edit.html", "list.html"))

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getFileList(dir string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if file.IsDir() || ext != ".txt" {
			continue
		}
		rep := regexp.MustCompile(`.txt$`)
		e := filepath.Base(rep.ReplaceAllString(file.Name(), ""))
		paths = append(paths, filepath.Join(dir, e))
	}
	return paths
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PathValue("title") // Go 1.22+
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PathValue("title") // Go 1.22+
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PathValue("title") // Go 1.22+
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	if err := p.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	files := getFileList("./")
	pl := &PageList{Pages: files}
	renderTemplate(w, "list.html", pl)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()

	// Go 1.22+ ルーティング
	mux.HandleFunc("GET /view/{title}", viewHandler)
	mux.HandleFunc("GET /edit/{title}", editHandler)
	mux.HandleFunc("POST /save/{title}", saveHandler)
	mux.HandleFunc("GET /", listHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
