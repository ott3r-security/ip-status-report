package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func readFile(fileName string) (string, error) {
	ipList, err := os.ReadFile(fileName)
	if err != nil {
		log.Printf("Error opening file. Needs to be named ip_list.txt: %v", err)
		return "", err
	}
	return string(ipList), nil
}

func pingIP(ipAddress string, results chan<- string) {
	cmd := exec.Command("ping", "-n", "1", ipAddress)
	status := cmd.Run()
	if status != nil {
		results <- "offline"
	} else {
		results <- "online"
	}

}

func main() {
	fileName := "ip_list.txt"
	var wg sync.WaitGroup
	results := make(chan string, 50)

	ipList, err := readFile(fileName)

	if err != nil {
		log.Fatalf("Error creating IP list: %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(ipList))
	for scanner.Scan() {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			pingIP(ip, results)
		}(scanner.Text())
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	outputFile, err := os.OpenFile("tester_status.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening output file:", err)
		return
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Include the date as the first cell in the row
	date := time.Now().Format("2006-01-02")
	row := []string{date}

	// Collect ping results
	for result := range results {
		row = append(row, result)
	}

	// Write the collected results to the CSV file
	if err := writer.Write(row); err != nil {
		fmt.Println("Error writing to CSV:", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
	}
}
