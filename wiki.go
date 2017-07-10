package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
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

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t := template.Must(template.ParseFiles(tmpl + ".html"))
	err := t.Execute(w, p)
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
	t := template.Must(template.ParseFiles("list.html"))
	files := dirwalk("./")
	pl := &PageList{Pages: files}
	err := t.Execute(w, pl)
	if err != nil {
		panic(err)
	}
}

//func init() {
func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/", listHandler)
	http.ListenAndServe(":8080", nil)
}
