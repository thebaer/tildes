package main

import (
	"os"
	"fmt"
	"time"
	"flag"
	"bufio"
	"strings"
	"path/filepath"
	"text/template"
)

var searchDir string
var searchDirs []string

func main() {
	fmt.Println("Starting...")

	// Get any arguments
	dirPtr := flag.String("d", "Code", "Directory to scan for each user.")
	flag.Parse()
	searchDir = *dirPtr
	searchDirs = strings.Split(searchDir, ",")

	if len(searchDirs) > 1 {
		generate(findProjects(), searchDirs[0])
	} else {
		generate(findProjects(), searchDir)
	}
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
	var files []string
	users := make(map[string]User)

	if len(searchDirs) > 1 {
		for _, d := range searchDirs {
			files, _ = filepath.Glob("/home/*/" + d + "/*")
			mapFiles(&files, users)
		}
	} else {
		files, _ = filepath.Glob("/home/*/" + searchDir + "/*")
		mapFiles(&files, users)
	}


	return users
}

func mapFiles(files *[]string, users map[string]User) {
	for _, path := range *files {
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
			fmt.Printf("Found %s for ~%s.\n", pparts[3], uname)
			projs := []Project{*proj}
			u = User{Name: uname, Projects: projs}
		} else {
			u.Projects = append(u.Projects, *proj)
		}
		users[uname] = u
	}
}

type Page struct {
	FolderName string
	Folders []string
	Host string
	Users map[string]User
	Updated string
	UpdatedForHumans string
}

func graphicalName(n string) string {
	r := strings.NewReplacer("tilde", "~", "ctrl-c", "^C", "nuclear", "&#9762;")
	return r.Replace(n)
}

func Split(s string, d string) []string {
	arr := strings.Split(s, d)
	return arr
}

func ListDirs(dirs []string) string {
	if len(dirs) > 1 {
		for i := 0; i < len(dirs); i++ {
			d := dirs[i]
			d = fmt.Sprintf("<strong>%s</strong>", d)
			dirs[i] = d
		}
		return strings.Join(dirs, " or ")
	}
	return fmt.Sprintf("<strong>%s</strong>", dirs[0])
}

func generate(users map[string]User, outputFile string) {
	fmt.Println("Generating page.")

	f, err := os.Create(os.Getenv("HOME") + "/public_html/" + strings.ToLower(outputFile) + ".html")
	if err != nil {
		panic(err)
	}
	
	defer f.Close()

	funcMap := template.FuncMap {
		"Split": Split,
		"ListDirs": ListDirs,
	}
	
	w := bufio.NewWriter(f)
	template, err := template.New("").Funcs(funcMap).ParseFiles("templates/code.html")
	if err != nil {
		panic(err)
	}

	// Extra page data
	host, _ := os.Hostname()
	curTime := time.Now().UTC()
	updatedReadable := curTime.Format(time.RFC1123)
	updated := curTime.Format(time.RFC3339)

	// Generate the page
	page := &Page{FolderName: searchDirs[0], Folders: searchDirs, Host: graphicalName(host), UpdatedForHumans: updatedReadable, Updated: updated, Users: users}
	template.ExecuteTemplate(w, "code", page)
	w.Flush()

	fmt.Println("DONE!")
}
