package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/francoganga/statement_to_csv/internal/parser"
)

var re = regexp.MustCompile(`(?m)^([0-9]{2}/[0-9]{2}/[0-9]{2})\s+([0-9]+)\s+(.*?)\s{2,}(.*?)\s{2,}(.*)\n(.*)`)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file1> <file2> ...")
		os.Exit(1)
	}

	filenames := os.Args[1:]

	file, err := os.Create("out.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	defer writer.Flush()

	header := []string{"Date", "Code", "Description", "Amount", "Balance"}

	csvData := [][]string{}

	writer.Write(header)

	for _, filename := range filenames {
		//open and read file
		fileContent, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		matches, err := GetMatchesFromFile(fileContent)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("len(matches)=%v\n", len(matches))

		for _, match := range matches {

			consumo, err := parser.New(match).Parse()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(consumo)

			t, err := time.Parse("02/01/06", consumo.Date)
			if err != nil {
				log.Fatal(err)
			}

			csvData = append(csvData, []string{t.Format("2006-01-02"), consumo.Code, consumo.Description, fmt.Sprintf("%d", consumo.Amount/100), fmt.Sprintf("%d", consumo.Balance/100)})
		}
		fmt.Printf("csvData=%v\n", csvData)

		for _, row := range csvData {
			if err := writer.Write(row); err != nil {
				panic(err)
			}

		}
	}
}

func GetMatchesFromFile(contents []byte) ([]string, error) {

	command := exec.Command("pdftotext", "-layout", "-f", "1", "-l", "3", "-", "-")

	stdin, err := command.StdinPipe()

	if err != nil {
		return nil, err
	}

	var outb bytes.Buffer

	command.Stdout = &outb

	if err = command.Start(); err != nil { //Use start, not run
		fmt.Println("An error occured: ", err) //replace with logger, or anything you want
	}

	_, err = io.WriteString(stdin, string(contents))

	if err != nil {
		return nil, err
	}

	stdin.Close()

	err = command.Wait()

	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`(?m)^([0-9]{2}/[0-9]{2}/[0-9]{2})\s+([0-9]+)\s+(.*?)\s{2,}(.*?)\s{2,}(.*)\n(.*)`)

	return re.FindAllString(outb.String(), -1), nil
}

