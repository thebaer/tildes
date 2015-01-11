package main

import (
	"fmt"
	"io/ioutil"
	"time"
	"strconv"
	"text/template"
	"os"
	"bufio"
	"regexp"
)

const entriesPath = "./entries/"
const templatesPath = "./templates/"
const outputPath = "./html/"

type Entry struct {
	Date string
	Body []byte
}

func loadEntry(rawDate string) (*Entry, error) {
	filename := entriesPath + rawDate
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Get raw date parts for formatting
	year := rawDate[:4]
	month, moErr := strconv.Atoi(rawDate[4:6])
	if moErr != nil {
		return nil, moErr
	}
	date, dateErr := strconv.Atoi(rawDate[6:8])
	if dateErr != nil {
		return nil, dateErr
	}

	formattedDate := fmt.Sprintf("%d %s %s", date, time.Month(month).String(), year)

	return &Entry{Date: formattedDate, Body: body}, nil
}

func main() {
	fmt.Println()
	fmt.Println("    ~log generator v1.0")
	fmt.Println()

	entryFiles := getEntries()
	entries := make([]Entry, len(*entryFiles))
	i := 0
	for _, file := range *entryFiles {
		entry, err := loadEntry(file)
		if err != nil {
			fmt.Printf("Error, skipping entry %s: %s\n", file, err)
			continue
		}
		fmt.Printf("Adding entry %s\n", file)
		entries[i] = *entry
		i++
	}
	
	generateLog(entries)
}

var validFileFormat = regexp.MustCompile("^[0-9]{8}$")

func getEntries() *[]string {
	files, _ := ioutil.ReadDir(entriesPath)
	fileList := make([]string, len(files))
	fileCount := 0
	// Traverse file list in reverse, i.e. newest to oldest
	for i := len(files)-1; i >= 0; i-- {
		file := files[i]
		if validFileFormat.Match([]byte(file.Name())) {
			fileList[fileCount] = file.Name()
			fileCount++
		}
	}
	fileList = fileList[:fileCount]
	return &fileList
}

func generateLog(entries []Entry) {
	file, err := os.Create(outputPath + "log.html")
	if err != nil {
		panic(err)
	}
	
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	template, err := template.ParseFiles(templatesPath + "log.html")
	if err != nil {
		panic(err)
	}
    template.Execute(writer, entries)
    writer.Flush()
}

