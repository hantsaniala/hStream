package gen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hantsaniala/hStream/pkg/utils"
)

func GenKey(destPath string) error {
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		os.MkdirAll(destPath, 0644)
	}

	encryptionKeyPath := filepath.Join(destPath, utils.GetEnv("KEY"))
	cmd := exec.Command("openssl", "rand", "-out", "-", "16")
	out, err := cmd.Output()
	if err != nil && err.Error() != "exit status 1" {
		return err
	}
	sOut := strings.TrimSpace(string(out))

	utils.WriteToFile(encryptionKeyPath, []string{sOut})
	return nil
}

func GenKeyinfo(destPath, keyURI string) error {
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		os.MkdirAll(destPath, 0644)
	}

	keyFilename := filepath.Join(destPath, utils.GetEnv("KEY"))
	keyInfoFile := filepath.Join(destPath, utils.GetEnv("KEYINFO"))
	cmd := exec.Command("openssl", "rand", "-hex", "16")
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	keyIV := strings.TrimSpace(string(out))

	keyInfoContent := []string{
		keyURI,
		keyFilename,
		keyIV,
	}

	utils.WriteToFile(keyInfoFile, keyInfoContent)
	return nil
}
