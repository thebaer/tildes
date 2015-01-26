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

const (
	searchDir = "Code"
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
	files, _ := filepath.Glob("/home/*/" + searchDir + "/*")
	users := make(map[string]User)

	for _, path := range files {
		pparts := strings.Split(path, "/")
		uname := pparts[2]
		fname := filepath.Base(path)

		// Exclude certain file names
		if strings.HasPrefix(fname, ".") || strings.HasSuffix(fname, ".log") || strings.HasSuffix(fname, "~") {
			continue
		}

		// Ensure file is other-readable
		// TODO: just detect if we can actually read this, instead
		info, _ := os.Stat(path)
		if info.Mode() & 0004 == 0 {
			continue
		}

		proj := &Project{Name: fname, Path: strings.Replace(path, "/home/", "~", -1)}
		u, exists := users[uname]
		if !exists {
			fmt.Printf("Found %s for ~%s.\n", searchDir, uname)
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
	FolderName string
	Host string
	Users map[string]User
	Updated string
	UpdatedForHumans string
}

func graphicalName(n string) string {
	r := strings.NewReplacer("tilde", "~", "ctrl-c", "^C", "nuclear", "&#9762;")
	return r.Replace(n)
}

func generate(users map[string]User) {
	fmt.Println("Generating page.")

	f, err := os.Create(os.Getenv("HOME") + "/public_html/" + strings.ToLower(searchDir) + ".html")
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
	curTime := time.Now().UTC()
	updatedReadable := curTime.Format(time.RFC1123)
	updated := curTime.Format(time.RFC3339)

	// Generate the page
	page := &Page{FolderName: searchDir, Host: graphicalName(host), UpdatedForHumans: updatedReadable, Updated: updated, Users: users}
	template.ExecuteTemplate(w, "code", page)
	w.Flush()

	fmt.Println("DONE!")
}
