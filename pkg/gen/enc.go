package gen

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hantsaniala/hStream/pkg/utils"
)

func GenKey(destPath string) error {
	encryptionKeyPath := filepath.Join(destPath, "enc.key")
	cmd := exec.Command("openssl", "rand", "16")
	out, err := cmd.Output()
	if err != nil && err.Error() != "exit status 1" {
		return err
	}

	utils.WriteToFile(encryptionKeyPath, []string{string(out)})
	return nil
}

func GenKeyinfo(destPath, keyURI string) error {
	keyFilename := "enc.key"
	keyInfoPath := filepath.Join(destPath, "enc.keyinfo")
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

	utils.WriteToFile(keyInfoPath, keyInfoContent)
	return nil
}
