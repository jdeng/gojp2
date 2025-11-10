package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"os"

	"github.com/jdeng/gojp2"
)

func main() {
	quality := flag.Int("quality", 90, "JPEG encoding quality (1-100)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <input.jp2> <output.jpg>\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	if *quality < 1 || *quality > 100 {
		log.Fatalf("invalid quality %d: must be between 1 and 100", *quality)
	}

	inputPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	input, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("open input: %v", err)
	}
	defer input.Close()

	img, err := gojp2.DecodeReader(input)
	if err != nil {
		log.Fatalf("decode jp2: %v", err)
	}

	output, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer func() {
		if cerr := output.Close(); cerr != nil {
			log.Printf("close output: %v", cerr)
		}
	}()

	if err := jpeg.Encode(output, img, &jpeg.Options{Quality: *quality}); err != nil {
		log.Fatalf("encode jpeg: %v", err)
	}
}
