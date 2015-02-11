package main

import (
	"encoding/json"
	"strings"
	"bufio"
	"os"
	"io/ioutil"
	"os/exec"
	"fmt"
	"regexp"
	"net/http"
	"flag"
	"time"
	"text/template"
)

func main() {
	// Get arguments
	outFilePtr := flag.String("f", "where", "Outputted HTML filename (without .html)")
	flag.Parse()

	// Get online users with `who`
	users := who()

	// Fetch user locations based on IP address
	for i := range users {
		getGeo(&users[i])
	}

	// Generate page
	generate(users, *outFilePtr)
}

type User struct {
	Name string
	IP string
	Region string
	Country string
	CurrentTime string
}

var ipRegex = regexp.MustCompile("(([0-9]{1,3}[.-]){3}[0-9]{1,3})")

func who() []User {
	fmt.Println("who --ips")

	cmd := exec.Command("who", "--ips")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}

	ips := make(map[string]string)

	r := bufio.NewReader(stdout)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineParts := strings.Split(scanner.Text(), " ")
		name := lineParts[0]

		fmt.Println(name)

		// Extract IP address
		ipMatch := ipRegex.FindAllString(scanner.Text(), 1)
		if len(ipMatch) == 0 {
			continue
		}

		// Normalize any host names with dashes
		newIp := strings.Replace(ipMatch[0], "-", ".", -1)

		ips[newIp] = name
	}

	users := make([]User, len(ips))
	i := 0
	for ip, name := range ips {
		users[i] = User{Name: name, IP: ip}
		i++
	}

	return users
}

func getTimeInZone(tz string) string {
	cmd := exec.Command("date", "+%A %H:%M")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TZ=%s", tz))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}

	r := bufio.NewReader(stdout)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func getGeo(u *User) {
	fmt.Printf("Fetching %s location...\n", u.Name)

    response, err := http.Get(fmt.Sprintf("https://freegeoip.net/json/%s", u.IP))
    if err != nil {
        fmt.Printf("%s", err)
        os.Exit(1)
    } else {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err != nil {
            fmt.Printf("%s", err)
            os.Exit(1)
        }

		var dat map[string]interface{}

		if err := json.Unmarshal(contents, &dat); err != nil {
			fmt.Println(err)
			return
		}
		region := dat["region_name"].(string)
		country := dat["country_name"].(string)

		u.CurrentTime = getTimeInZone(dat["time_zone"].(string))
		u.Region = region
		u.Country = country
    }
}

func prettyLocation(region, country string) string {
	if region != "" {
		return fmt.Sprintf("%s, %s", region, country)
	}
	return country
}

type Page struct {
	Users []User
	Updated string
	UpdatedForHumans string
}

func generate(users []User, outputFile string) {
	fmt.Println("Generating page.")

	f, err := os.Create(os.Getenv("HOME") + "/public_html/" + strings.ToLower(outputFile) + ".html")
	if err != nil {
		panic(err)
	}
	
	defer f.Close()

	funcMap := template.FuncMap {
		"Location": prettyLocation,
	}
	
	w := bufio.NewWriter(f)
	template, err := template.New("").Funcs(funcMap).ParseFiles("../templates/where.html")
	if err != nil {
		panic(err)
	}

	// Extra page data
	curTime := time.Now().UTC()
	updatedReadable := curTime.Format(time.RFC1123)
	updated := curTime.Format(time.RFC3339)

	// Generate the page
	page := &Page{Users: users, UpdatedForHumans: updatedReadable, Updated: updated}
	template.ExecuteTemplate(w, "where", page)
	w.Flush()

	fmt.Println("DONE!")
}
