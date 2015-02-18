package main

import (
	"os"
	"os/exec"
	"fmt"
	"time"
	"flag"
	"sort"
	"bufio"
	"strconv"
	"strings"
	"io/ioutil"
	"text/template"

	"github.com/thebaer/tildes/store"
)

var (
	scoresPath = "/home/krowbar/Code/irc/tildescores.txt"
	jackpotPath = "/home/krowbar/Code/irc/tildejackpot.txt"
	addictionData = "/home/karlen/bin/tilderoyale"
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
		jackpotPath = "/home/bear/tildejackpot.txt"
		addictionData = "/home/bear/addicted.sh"
	}

	headers := []string{ "User", "Tildes", "Last Collected", "Addiction", "# Asks", "Avg.", "Last Amt." }

	scoresData := store.ReadRows(scoresPath, "&^%")
	updatesData := store.ReadRows(scoreDeltasPath, deltaDelimiter)

	scoresData = checkScoreDelta(scoresData, updatesData)
	scoresTable := buildScoresTable(scoresData, headers)

	generate("!tilde scores", getFile(jackpotPath), sortScore(scoresTable), *outPtr)
}

type Table struct {
	Headers []string
	Rows []store.Row
}

type By func(r1, r2 *store.Row) bool
func (by By) Sort(rows []store.Row) {
	rs := &rowSorter {
		rows: rows,
		by: by,
	}
	sort.Sort(rs)
}
type rowSorter struct {
	rows []store.Row
	by func(r1, r2 *store.Row) bool
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
	score := func(r1, r2 *store.Row) bool {
		s1, _ := strconv.Atoi(r1.Data[1])
		s2, _ := strconv.Atoi(r2.Data[1])
		return s1 < s2
	}
	decScore := func(r1, r2 *store.Row) bool {
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

func trimTrailingZeros(n float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", n), "0"), ".")
}

func trimTrailingZerosShort(n float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", n), "0"), ".")
}

func niceTime(sec int) string {
	if sec == 0 {
		return "-"
	}

	if sec < 60 {
		return fmt.Sprintf("%dsec", sec)
	} else if sec < 3600 {
		return fmt.Sprintf("%smin", trimTrailingZerosShort(float64(sec) / 60.0))
	} else if sec < 86400 {
		return fmt.Sprintf("%shr", trimTrailingZerosShort(float64(sec) / 3600.0))
	} else {
		return fmt.Sprintf("%sdy", trimTrailingZerosShort(float64(sec) / 86400.0))
	}
}

type LastScore struct {
	LastUpdate int
	LastScore int
	LastIncrement int
	Times int
	ScoreOffset int
	Addiction int
}

func checkScoreDelta(scoreRows, deltaRows *[]store.Row) *[]store.Row {
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

		users[r.Data[0]] = LastScore{ LastScore: score, LastUpdate: update, LastIncrement: inc, Times: times, ScoreOffset: so } 
	}

	// Fetch IRC log data
	fmt.Println("Fetching IRC log data")
	cmd := exec.Command(addictionData, "/home/karlen/bin/addictedtotilde -c")
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
		logRow := strings.Split(scanner.Text(), "\t")

		uname := strings.TrimSpace(logRow[0])
		asks, _ := strconv.Atoi(strings.TrimSpace(logRow[3]))
		addiction, _ := strconv.Atoi(strings.TrimSpace(logRow[1]))

		u, exists := users[uname]
		if !exists {
			u = LastScore{ ScoreOffset: 0 }
		}
		u.ScoreOffset = 0
		u.Times = asks
		u.Addiction = addiction

		fmt.Println(fmt.Sprintf("%d", u.Times))
		users[uname] = u
	}

	// Add in scores data
	fmt.Println(fmt.Sprintf("Reading scores data on %d users", len(*scoreRows)))
	for i := range *scoreRows {
		r := (*scoreRows)[i]
		u, exists := users[r.Data[0]]
		fmt.Println(u)

		score, _ := strconv.Atoi(r.Data[1])
		update, _ := strconv.Atoi(r.Data[2])

		// Fill in any missing users
		if !exists {
			u = LastScore{ LastScore: score, LastIncrement: -1, LastUpdate: update, Times: 0, ScoreOffset: score, Addiction: 0 }
			users[r.Data[0]] = u
		}

		// Match up "last collection" with rest of table data
		if update > u.LastUpdate {
			u.LastIncrement = score - u.LastScore
			u.LastUpdate = update
			u.LastScore = score
			u.Times++
		}

		r.Data = append(r.Data, niceTime(u.Addiction))
		
		var asksStr string
		if u.Times > 0 {
			asksStr = strconv.Itoa(u.Times)
		} else {
			asksStr = "-"
		}
		r.Data = append(r.Data, asksStr)

		var avgStr string
		if u.Times > 0 {
			avg := float64(score - u.ScoreOffset) / float64(u.Times)
			avgStr = trimTrailingZeros(avg)
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
		userData := fmt.Sprintf("%d%s%d%s%d%s%d%s%d", v.LastScore, deltaDelimiter, v.LastUpdate, deltaDelimiter, v.LastIncrement, deltaDelimiter, v.Times, deltaDelimiter, v.ScoreOffset)
		_, err = f.WriteString(fmt.Sprintf("%s%s%s\n", k, deltaDelimiter, userData))
		if err != nil {
			fmt.Println(err)
		}
	}

	return scoreRows
}

func buildScoresTable(rows *[]store.Row, headers []string) *Table {
	t := &Table{Headers: headers, Rows: nil}

	const layout = "Jan 2, 2006 3:04pm MST"
	for i, r := range *rows {
		data := r.Data
		time := parseTimestamp(r.Data[2])
		r.Data[2] = time.UTC().Format(layout)
		outRow := &store.Row{Data: data}
		(*rows)[i] = *outRow
	}
	t.Rows = *rows

	return t
}

func getFile(path string) string {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(f)
}

type Page struct {
	Title string
	Table Table
	Updated string
	UpdatedForHumans string
	Jackpot int
}

func add(x, y int) int {
	return x + y
}

func generate(title, jackpot string, table *Table, outputFile string) {
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
	template, err := template.New("").Funcs(funcMap).ParseFiles("../templates/scores.html")
	if err != nil {
		panic(err)
	}

	// Extra page data
	curTime := time.Now().UTC()
	updatedReadable := curTime.Format(time.RFC1123)
	updated := curTime.Format(time.RFC3339)

	// Jackpot parsing
	jp, err := strconv.Atoi(jackpot)
	if err != nil {
		fmt.Println(err)
		jp = -1
	}

	// Generate the page
	page := &Page{Title: title, Table: *table, UpdatedForHumans: updatedReadable, Updated: updated, Jackpot: jp}
	template.ExecuteTemplate(w, "table", page)
	w.Flush()

	fmt.Println("DONE!")
}
