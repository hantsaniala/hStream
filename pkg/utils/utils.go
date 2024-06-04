package utils

import (
	"fmt"
	"os"
)

func WriteToFile(filename string, lines []string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file for writing", err)
		return
	}
	defer f.Close()

	for _, line := range lines {
		_, err := f.WriteString(line + "\n")
		if err != nil {
			fmt.Println("Error writing to file", err)
			return
		}
	}
}
