package main

import (
    "github.com/boriska70/tugbot-rs-es/junitxml"
    "html/template"
    "io/ioutil"
    "net/http"
    "regexp"
    "errors"
    "fmt"
)

type Page struct {
    Title string
    Body  []byte
}

var templates = template.Must(template.ParseFiles("templates/edit.html", "templates/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var dataPath = "data/"

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m:=validPath.FindStringSubmatch(r.URL.Path)
    if m== nil {
        http.NotFound(w,r)
        return "", errors.New("Invalid Page Title")
    }
    return m[2], nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    /*    t, err := template.ParseFiles("templates/edit.html", "templates/view.html")
	if err != nil {
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	    return
		err = t.Execute(w, p)
	}*/
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(dataPath+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(dataPath+filename)
    if err != nil {
	return nil, err
    }
    return &Page{Title:title, Body:body}, nil
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
//    title := r.URL.Path[len("/view/"):]
    p, err := loadPage(title)
    if err != nil {
	http.Redirect(w, r, "/edit/" + title, http.StatusFound)
	return
    }
    renderTemplate(w, "view", p)
    //    fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
//    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
	p = &Page{Title:title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    //title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title:title, Body: []byte(body)}
    err := p.save()
    if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }
    http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func postDataHandler(w http.ResponseWriter, r *http.Request)  {
    w.WriteHeader(http.StatusOK)
    b, _:=ioutil.ReadAll(r.Body)
    w.Write([]byte(junitxml.HandleJUnitXml(b)));



    return

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w,r,"/view/FrontPage", http.StatusFound)
    return
}

func main() {

    defer func() {fmt.Println("Server is up and running")}()

    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/post/", postDataHandler)
//    http.HandleFunc("/", rootHandler)
    fmt.Println(http.ListenAndServe(":8080", nil))
    fmt.Println("aaaaaaaaaaaaaaaaa")
}