package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

type PageList struct {
	Pages []string
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
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

const lenPath = len("/view/")

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[lenPath:]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[lenPath:]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t := template.Must(template.ParseFiles(tmpl + ".html"))
	err := t.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[lenPath:]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	files := dirwalk("./")
	pl := &PageList{Pages: files}
	renderTemplate(w, "list", pl)
}

//func init() {
func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/", listHandler)
	http.ListenAndServe(":8080", nil)
}
