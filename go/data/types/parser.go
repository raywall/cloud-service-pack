package types

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseStringToJSON(content string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("error when analyzing JSON: %w", err)
	}
	return data, nil
}

func ParseStringToYAML(content string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, fmt.Errorf("error when analyzing yaml: %w", err)
	}
	return data, nil
}

func ParseStringToCSV(content string) (interface{}, error) {
	data := make([]map[string]interface{}, 0, 0)

	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error when analyzing CSV: %w", err)
	}

	if len(records) < 2 {
		return records, nil
	}

	// csv to map converter
	headers := records[0]

	for i := 1; i < len(records); i++ {
		row := make(map[string]interface{})
		for j := 0; j < len(headers) && j < len(records[i]); j++ {
			row[headers[j]] = records[i][j]
		}
		data = append(data, row)
	}
	return data, nil
}
