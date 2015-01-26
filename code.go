package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"strings"
	"path/filepath"
	"text/template"
)

func main() {
	fmt.Println("Starting...")
	generate(findProjects())
}

type User struct {
	Name string
	Projects []Project
}

type Project struct {
	Name string
	Path string
}

func findProjects() map[string]User {
	files, _ := filepath.Glob("/home/*/Code/*")
	users := make(map[string]User)

	for _, path := range files {
		pparts := strings.Split(path, "/")
		uname := pparts[2]
		proj := &Project{Name: filepath.Base(path), Path: strings.Replace(path, "/home/", "~", -1)}
		u, exists := users[uname]
		if !exists {
			fmt.Printf("Found Code for ~%s.\n", uname)
			projs := []Project{*proj}
			u = User{Name: uname, Projects: projs}
		} else {
			u.Projects = append(u.Projects, *proj)
		}
		users[uname] = u
	}
	return users
}

type Page struct {
	Host string
	Users map[string]User
	Updated string
	UpdatedForHumans string
}

func generate(users map[string]User) {
	fmt.Println("Generating page.")

	f, err := os.Create(os.Getenv("HOME") + "/public_html/code.html")
	if err != nil {
		panic(err)
	}
	
	defer f.Close()
	
	w := bufio.NewWriter(f)
	template, err := template.ParseFiles("templates/code.html")
	if err != nil {
		panic(err)
	}

	// Extra page data
	host, _ := os.Hostname()
	curTime := time.Now()
	updatedReadable := curTime.UTC().Format(time.RFC1123)
	updated := curTime.UTC().Format(time.RFC3339)

	// Generate the page
	page := &Page{Host: host, UpdatedForHumans: updatedReadable, Updated: updated, Users: users}
	template.ExecuteTemplate(w, "code", page)
	w.Flush()

	fmt.Println("DONE!")
}
