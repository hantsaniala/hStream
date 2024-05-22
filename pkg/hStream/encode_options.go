package hStream

type EncodeOptions struct {
	VideoBitrate string
	VideoMaxRate string
	VideoMinRate string
	VideoBufSize string
	AudioBitrate string
}

func GetEncodeOption() map[int]EncodeOptions {
	return map[int]EncodeOptions{
		1080: {
			"5M",
			"5M",
			"5M",
			"10M",
			"96K",
		},
		720: {
			"3M",
			"3M",
			"3M",
			"3M",
			"96K",
		},
		540: {
			"2M",
			"2M",
			"2M",
			"2M",
			"48K",
		},
		360: {
			"1M",
			"1M",
			"1M",
			"1M",
			"48K",
		},
	}
}
