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

var (
	scoresPath = "/home/krowbar/Code/irc/tildescores.txt"
)

const (
	scoreDeltasPath = "/home/bear/scoredeltas.txt"
	deltaDelimiter = "+++"
)

func main() {
	fmt.Println("Starting...")

	// Get any arguments
	outPtr := flag.String("o", "tildescores", "Output file name")
	isTestPtr := flag.Bool("t", false, "Specifies we're developing")
	flag.Parse()

	if *isTestPtr {
		scoresPath = "/home/bear/tildescores.txt"
	}

	headers := []string{ "User", "Tildes", "Last Collected", "# Asks", "Avg.", "Last Amt." }

	scoresData := readData(scoresPath, "&^%")
	updatesData := readData(scoreDeltasPath, deltaDelimiter)

	scoresData = checkScoreDelta(scoresData, updatesData)
	scoresTable := buildScoresTable(scoresData, headers)

	generate("tilde collectors", sortScore(scoresTable), *outPtr)
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

type LastScore struct {
	LastUpdate int
	LastScore int
	LastIncrement int
	Times int
	ScoreOffset int
	TimesOffset int
}

func checkScoreDelta(scoreRows *[]Row, deltaRows *[]Row) *[]Row {
	users := make(map[string]LastScore)

	// Read score delta data
	for i := range *deltaRows {
		r := (*deltaRows)[i]

		if len(r.Data) < 4 {
			break
		}

		score, _ := strconv.Atoi(r.Data[1])
		update, _ := strconv.Atoi(r.Data[2])
		inc, _ := strconv.Atoi(r.Data[3])
		times, _ := strconv.Atoi(r.Data[4])
		so, _ := strconv.Atoi(r.Data[5])
		to, _ := strconv.Atoi(r.Data[6])

		users[r.Data[0]] = LastScore{ LastScore: score, LastUpdate: update, LastIncrement: inc, Times: times, ScoreOffset: so, TimesOffset: to } 
	}

	for i := range *scoreRows {
		r := (*scoreRows)[i]
		u, exists := users[r.Data[0]]

		score, _ := strconv.Atoi(r.Data[1])
		update, _ := strconv.Atoi(r.Data[2])

		// Fill in any missing users
		if !exists {
			u = LastScore{ LastScore: score, LastIncrement: -1, LastUpdate: update, Times: 0, ScoreOffset: score, TimesOffset: 0 }
			users[r.Data[0]] = u
		}

		// Match up "last collection" with rest of table data
		if update > u.LastUpdate {
			u.LastIncrement = score - u.LastScore
			u.LastUpdate = update
			u.LastScore = score
			u.Times++
		}
		
		var asksStr string
		if u.Times > 0 {
			asksStr = strconv.Itoa(u.Times)
		} else {
			asksStr = "-"
		}
		r.Data = append(r.Data, asksStr)

		var avgStr string
		if u.Times > 0 {
			var avg float64 = float64(u.LastScore - u.ScoreOffset) / float64(u.Times)
			avgStr = strconv.FormatFloat(avg, 'f', -1, 32)
		} else {
			avgStr = "-"
		}
		r.Data = append(r.Data, avgStr)

		var lastIncStr string
		if u.LastIncrement > -1 {
			lastIncStr = strconv.Itoa(u.LastIncrement)
		} else {
			lastIncStr = "-"
		}
		r.Data = append(r.Data, lastIncStr)

		users[r.Data[0]] = u
		(*scoreRows)[i] = r
	}

	// Write deltas
	f, err := os.OpenFile(scoreDeltasPath, os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	for k, v := range users {
		userData := fmt.Sprintf("%d%s%d%s%d%s%d%s%d%s%d", v.LastScore, deltaDelimiter, v.LastUpdate, deltaDelimiter, v.LastIncrement, deltaDelimiter, v.Times, deltaDelimiter, v.ScoreOffset, deltaDelimiter, v.TimesOffset)
		_, err = f.WriteString(fmt.Sprintf("%s%s%s\n", k, deltaDelimiter, userData))
		if err != nil {
			fmt.Println(err)
		}
	}

	return scoreRows
}

func buildScoresTable(rows *[]Row, headers []string) *Table {
	table := &Table{Headers: headers, Rows: nil}

	const layout = "Jan 2, 2006 3:04pm MST"
	for i, r := range *rows {
		data := r.Data
		t := parseTimestamp(r.Data[2])
		r.Data[2] = t.UTC().Format(layout)
		row := &Row{Data: data}
		(*rows)[i] = *row
	}
	table.Rows = *rows

	return table
}

func readData(path string, delimiter string) *[]Row {
	f, _ := os.Open(path)
	
	defer f.Close()

	rows := []Row{}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		data := strings.Split(s.Text(), delimiter)
		row := &Row{Data: data}
		rows = append(rows, *row)
	}

	return &rows
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
