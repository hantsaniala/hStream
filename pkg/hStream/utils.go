package hStream

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/hantsaniala/hStream/pkg/utils"
)

func getMediaBase(mId string) string {
	mediaRoot := utils.GetEnv("MEDIA_ROOT")
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

func CopyFile(sourcePath, destPath string) error {
	// Source: https://stackoverflow.com/a/35353594/5527968

	srcFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// creates if file doesn't exist
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// check first var for number of bytes copied
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

// Generate random string
func RandomString(n int) string {
	// Source: https://golangdocs.com/generate-random-string-in-golang

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
