package gen

import (
	"os"
	"os/exec"
	"path/filepath"
)

func GenKeyPair(destPath string) error {
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		os.MkdirAll(destPath, 0644)
	}

	privateKeyPath := filepath.Join(destPath, "private_key.pem")
	publicKeyPath := filepath.Join(destPath, "public_key.pem")

	cmd := exec.Command("openssl",
		"genpkey",
		"-algorithm", "RSA",
		"-out", privateKeyPath,
		"-pkeyopt", "rsa_keygen_bits:2048",
	)

	_, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("openssl",
		"rsa",
		"-in", privateKeyPath,
		"-pubout",
		"-out", publicKeyPath,
	)

	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}
