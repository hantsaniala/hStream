package hStream

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getMediaBase(mId string) string {
	mediaRoot := GetEnv("MEDIA_ROOT")
	return fmt.Sprintf("%s/%s", mediaRoot, mId)
}

func getFileExt(filename string) string {
	s := strings.Split(filename, ".")
	return s[len(s)-1]

}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("couldn't copy to dest from source: %v", err)
	}

	inputFile.Close() // for Windows, close before trying to remove: https://stackoverflow.com/a/64943554/246801

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't remove source file: %v", err)
	}
	return nil
}
