package store

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"io/ioutil"
)

type Row struct {
	Data []string
}

func ReadData(path string) []byte {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return []byte(``)
	}

	return d
}

func ReadRows(path, delimiter string) *[]Row {
	f, _ := os.Open(path)
	
	defer f.Close()

	rows := []Row{}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		data := strings.Split(s.Text(), delimiter)
		r := &Row{Data: data}
		rows = append(rows, *r)
	}

	return &rows
}

func WriteData(path string, data []byte) {
	f, err := os.OpenFile(path, os.O_CREATE | os.O_RDWR | os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		fmt.Println(err)
	}
}

func WriteRows(path string, rows *[]Row, delimeter string) {
	f, err := os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	for _, r := range *rows {
		_, err = f.WriteString(strings.Join(r.Data, delimeter))
		if err != nil {
			fmt.Println(err)
		}
	}
}
