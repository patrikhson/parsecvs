package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Struct to represent a CSV row
type Record struct {
	Data map[string]string
}

func main() {
	// Command-line arguments
	fileName := flag.String("file", "", "Path to the CSV file")
	filter := flag.String("filter", "", `Filters in the format: "or(Företag,Kalle;Företag,Olle)" or "Företag,Kalle;Stad,Stockholm"`)
	selectFields := flag.String("select-fields", "", "Comma-separated list of field names to print (optional, defaults to all fields)")
	unique := flag.Bool("unique", false, "Ensure output lines are unique")
	listFields := flag.Bool("list-fields", false, "List available fields in the CSV file")
	flag.Parse()

	// Ensure a file is provided
	if *fileName == "" {
		fmt.Println("Usage: go run main.go -file=filename.csv [-list-fields] [-filter='or(Företag,Kalle;Företag,Olle)'] [-select-fields=field1,field2,...] [-unique]")
		os.Exit(1)
	}

	// Open and read the CSV file
	file, err := os.Open(*fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Ensure the file has content
	if len(records) < 2 {
		fmt.Println("The CSV file must contain a header and at least one data row.")
		os.Exit(1)
	}

	// Extract header (field names)
	header := records[0]
	fieldSet := make(map[string]bool)
	for _, field := range header {
		fieldSet[field] = true
	}

	// If the user requested to list fields, print them and exit
	if *listFields {
		fmt.Println("Available fields:")
		for _, field := range header {
			fmt.Println(field)
		}
		os.Exit(0)
	}

	// Validate `-select-fields` argument
	var fieldsToPrint []string
	if *selectFields == "" {
		fieldsToPrint = header
	} else {
		fieldsToPrint = strings.Split(*selectFields, ",")
		for _, field := range fieldsToPrint {
			if !fieldSet[field] {
				fmt.Printf("Error: Field '%s' does not exist in the CSV file.\n", field)
				os.Exit(2)
			}
		}
	}

	// Parse the filters
	andFilters, orFilters := parseFilters(*filter, fieldSet)

	// Create a slice of structs to store records
	var data []Record
	for _, row := range records[1:] {
		record := Record{Data: make(map[string]string)}
		for i, value := range row {
			record.Data[header[i]] = value
		}
		data = append(data, record)
	}

	// Process filtering
	filteredRecords := filterRecords(data, andFilters, orFilters)

	// Print selected fields
	printedLines := make(map[string]bool)
	for _, record := range filteredRecords {
		line := formatSelectedFields(record, fieldsToPrint)
		if *unique {
			if printedLines[line] {
				continue
			}
			printedLines[line] = true
		}
		fmt.Println(line)
	}
}

// Parses filters into AND and OR conditions, with field validation
func parseFilters(filter string, fieldSet map[string]bool) (map[string]string, map[string][]string) {
	andFilters := make(map[string]string)
	orFilters := make(map[string][]string)

	if filter == "" {
		return andFilters, orFilters
	}

	// Check for OR conditions
	orRegex := regexp.MustCompile(`or\(([^)]+)\)`)
	orMatches := orRegex.FindStringSubmatch(filter)
	if len(orMatches) > 1 {
		orParts := strings.Split(orMatches[1], ";")
		for _, part := range orParts {
			kv := strings.SplitN(part, ",", 2)
			if len(kv) == 2 {
				field, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
				if !fieldSet[field] {
					fmt.Printf("Error: Field '%s' does not exist in the CSV file.\n", field)
					os.Exit(2)
				}
				orFilters[field] = append(orFilters[field], value)
			}
		}
		// Remove the OR condition from the main filter string
		filter = orRegex.ReplaceAllString(filter, "")
	}

	// Process AND conditions
	parts := strings.Split(filter, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, ",", 2)
		if len(kv) == 2 {
			field, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			if !fieldSet[field] {
				fmt.Printf("Error: Field '%s' does not exist in the CSV file.\n", field)
				os.Exit(2)
			}
			andFilters[field] = value
		}
	}

	return andFilters, orFilters
}

// Filters records based on AND and OR conditions
func filterRecords(data []Record, andFilters map[string]string, orFilters map[string][]string) []Record {
	var filtered []Record

	for _, record := range data {
		match := true

		// Check AND filters
		for field, value := range andFilters {
			if record.Data[field] != value {
				match = false
				break
			}
		}

		// Check OR filters (only if the AND conditions matched)
		if match && len(orFilters) > 0 {
			orMatch := false
			for field, values := range orFilters {
				for _, val := range values {
					if record.Data[field] == val {
						orMatch = true
						break
					}
				}
			}
			match = orMatch // Ensure at least one OR condition is met
		}

		if match {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// Formats selected fields of a record into a string
func formatSelectedFields(record Record, fields []string) string {
	var output []string
	for _, field := range fields {
		output = append(output, record.Data[field])
	}
	return strings.Join(output, ", ")
}
