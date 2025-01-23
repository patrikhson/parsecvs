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
	filter := flag.String("filter", "", `Filters in the format: "name1,value1 name2,'value with space' 'name 3','value 3'"`)
	selectFields := flag.String("select-fields", "", "Comma-separated list of field names to print (optional, defaults to all fields)")
	unique := flag.Bool("unique", false, "Ensure output lines are unique")
	flag.Parse()

	// Check if the required arguments are provided
	if *fileName == "" {
		fmt.Println("Usage: go run main.go -file=filename.csv -filter='name1,value1 name2,value2' [-select-fields=field1,field2,...] [-unique]")
		os.Exit(1)
	}

	// Parse the filters (empty if not provided)
	filters := parseFilters(*filter)

	// Open the CSV file
	file, err := os.Open(*fileName)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read the CSV file
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

	// Parse the header
	header := records[0]

	// Create a slice of structs to hold the data
	var data []Record
	for _, row := range records[1:] {
		record := Record{Data: make(map[string]string)}
		for i, value := range row {
			record.Data[header[i]] = value
		}
		data = append(data, record)
	}

	// Process and filter rows in memory
	filteredRecords := filterRecords(data, filters)

	// Determine the fields to print
	var fieldsToPrint []string
	if *selectFields == "" {
		// Default to printing all fields if -select-fields is not provided
		fieldsToPrint = header
	} else {
		fieldsToPrint = strings.Split(*selectFields, ",")
	}

	// Print selected fields from the filtered records
	printedLines := make(map[string]bool) // To track printed lines if -unique is enabled

	for _, record := range filteredRecords {
		line := formatSelectedFields(record, fieldsToPrint)

		if *unique {
			if printedLines[line] {
				continue // Skip duplicate lines
			}
			printedLines[line] = true
		}

		fmt.Println(line)
	}
}

// Parses the --filter argument into a map of field-value pairs
func parseFilters(filter string) map[string]string {
	filters := make(map[string]string)
	if filter == "" {
		return filters
	}

	// Regular expression to match key-value pairs with optional quotes
	re := regexp.MustCompile(`([^,]+),(?:"([^"]+)"|'([^']+)'|([^ ]+))`)

	matches := re.FindAllStringSubmatch(filter, -1)
	if len(matches) == 0 {
		fmt.Printf("Invalid filter format: %s. Ensure key-value pairs are in the format 'key,value'.\n", filter)
		os.Exit(1)
	}

	// Process matches
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2] + match[3] + match[4]) // Combine the matched value
		filters[key] = value
	}

	return filters
}

// Filters records based on the provided filters map
func filterRecords(data []Record, filters map[string]string) []Record {
	// If no filters are provided, return all records
	if len(filters) == 0 {
		return data
	}

	var filtered []Record
	for _, record := range data {
		matches := true
		for field, value := range filters {
			if record.Data[field] != value {
				matches = false
				break
			}
		}
		if matches {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// Formats selected fields of a record into a single string
func formatSelectedFields(record Record, fields []string) string {
	var output []string
	for _, field := range fields {
		output = append(output, record.Data[field])
	}
	return strings.Join(output, ", ")
}
