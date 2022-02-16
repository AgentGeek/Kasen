package main

import (
	"os"
	"strconv"

	"github.com/h2non/bimg"
	"github.com/rs1703/logger"
)

func main() {
	args := os.Args
	if len(args) <= 1 {
		return
	}

	input := args[1]
	width, _ := strconv.Atoi(args[2])
	height, _ := strconv.Atoi(args[3])
	crop := args[4]

	buf, err := os.ReadFile(input)
	if err != nil {
		logger.Err.Fatalln(err)
		return
	}

	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)

	jpeg, err := bimg.NewImage(buf).Convert(bimg.JPEG)
	if err != nil {
		logger.Err.Fatalln(err)
		return
	}

	out, err := bimg.NewImage(jpeg).Process(bimg.Options{
		Width:         width,
		Height:        height,
		StripMetadata: true,
		Crop:          crop == "true",
		Quality:       85,
		Interlace:     true,
		Interpolator:  bimg.Bicubic,
	})

	if err != nil {
		logger.Err.Fatalln(err)
		return
	}

	os.Stdout.Write(out)
}
