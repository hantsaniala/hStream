package hStream

import "fmt"

func getMediaBase(mId string) string {
	mediaRoot := GetEnv("MEDIA_ROOT")
	return fmt.Sprintf("%s/%s", mediaRoot, mId)
}
