package utils

import (
	"fmt"
	"os"
)

func WriteToFile(filename string, lines []string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("Error opening file for writing", err)
		return
	}
	defer f.Close()

	for i, line := range lines {
		if i > 0 {
			line = fmt.Sprintf("\n%s", line)
		}

		_, err := f.WriteString(line)
		if err != nil {
			fmt.Println("Error writing to file", err)
			return
		}
	}
}
