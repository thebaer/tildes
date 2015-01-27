package main

import (
	"os"
	"fmt"
	"time"
	"flag"
	"sort"
	"bufio"
	"strconv"
	"strings"
	"text/template"
)

const scores = "/home/krowbar/Code/irc/tildescores.txt"

func main() {
	fmt.Println("Starting...")

	// Get any arguments
	outPtr := flag.String("o", "tildescores", "Output file name")
	flag.Parse()

	headers := []string{ "User", "Tildes", "Last Collection" }
	generate("tilde collectors", sortScore(readData(scores, "&^%", headers)), *outPtr)
}

type Table struct {
	Headers []string
	Rows []Row
}

type Row struct {
	Data []string
}

type By func(r1, r2 *Row) bool
func (by By) Sort(rows []Row) {
	rs := &rowSorter {
		rows: rows,
		by: by,
	}
	sort.Sort(rs)
}
type rowSorter struct {
	rows []Row
	by func(r1, r2 *Row) bool
}
func (r *rowSorter) Len() int {
	return len(r.rows)
}
func (r *rowSorter) Swap(i, j int) {
	r.rows[i], r.rows[j] = r.rows[j], r.rows[i]
}
func (r *rowSorter) Less(i, j int) bool {
	return r.by(&r.rows[i], &r.rows[j])
}

func sortScore(table *Table) *Table {
	score := func(r1, r2 *Row) bool {
		s1, _ := strconv.Atoi(r1.Data[1])
		s2, _ := strconv.Atoi(r2.Data[1])
		return s1 < s2
	}
	decScore := func(r1, r2 *Row) bool {
		return !score(r1, r2)
	}
	By(decScore).Sort(table.Rows)

	return table
}

func parseTimestamp(ts string) time.Time {
	t, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Unix(t, 0)
}

func readData(path string, delimiter string, headers []string) *Table {
	f, _ := os.Open(path)
	
	defer f.Close()

	rows := []Row{}
	table := &Table{Headers: headers, Rows: nil}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	const layout = "Jan 2, 2006 3:04pm MST"
	for s.Scan() {
		data := strings.Split(s.Text(), delimiter)
		t := parseTimestamp(data[2])
		data[2] = t.UTC().Format(layout)
		row := &Row{Data: data}
		rows = append(rows, *row)
	}
	table.Rows = rows

	return table
}

type Page struct {
	Title string
	Table Table
	Updated string
	UpdatedForHumans string
}

func add(x, y int) int {
	return x + y
}

func generate(title string, table *Table, outputFile string) {
	fmt.Println("Generating page.")

	f, err := os.Create(os.Getenv("HOME") + "/public_html/" + outputFile + ".html")
	if err != nil {
		panic(err)
	}
	
	defer f.Close()

	funcMap := template.FuncMap {
		"add": add,
	}

	w := bufio.NewWriter(f)
	template, err := template.New("").Funcs(funcMap).ParseFiles("templates/table.html")
	if err != nil {
		panic(err)
	}

	// Extra page data
	curTime := time.Now().UTC()
	updatedReadable := curTime.Format(time.RFC1123)
	updated := curTime.Format(time.RFC3339)

	// Generate the page
	page := &Page{Title: title, Table: *table, UpdatedForHumans: updatedReadable, Updated: updated}
	template.ExecuteTemplate(w, "table", page)
	w.Flush()

	fmt.Println("DONE!")
}
