package hStream

import (
	"fmt"
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
