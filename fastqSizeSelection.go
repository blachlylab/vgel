package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"os"
)

func readGzFile(filename, filenameout string, min, max int) error {
	// read the gzipped input file
	fi, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fi.Close()
	fz, err := gzip.NewReader(fi)
	if err != nil {
		return err
	}
	defer fz.Close()

	// init writing the gzipped output file
	fo, err := os.Create(filenameout)
	if err != nil {
		return err
	}
	defer fo.Close()
	fzo := gzip.NewWriter(fo)
	if err != nil {
		return err
	}
	defer fzo.Close()

	scanner := bufio.NewScanner(fz)
	for scanner.Scan() {
		/* fastq format :
			1: header line
			2: sequence line
			3: header line (may only be "+")
			4: phred quality score line
		must parse through the file 4 lines at a time
		*/
		header := scanner.Text()
		scanner.Scan()
		sequence := scanner.Text()
		scanner.Scan()
		bonus := scanner.Text()
		scanner.Scan()
		quality := scanner.Text()
		// check if this read meets our size requirements
		if len(sequence) >= min && len(sequence) <= max {
			for _, line := range []string{header, sequence, bonus, quality} {
				//fmt.Println(line)
				fzo.Write([]byte(line + "\n"))
			}
		}
	}
	return nil
}

func main() {
	flagFastq := flag.String("fastq", "", "fastq.gz sequence file")
	flagOut := flag.String("out", "output.fastq.gz", "output file name")
	flagMax := flag.Int("max", 1000, "max read length size")
	flagMin := flag.Int("min", 0, "min read length size")
	flag.Parse()
	// maybe do arg checking someday

	if _, err := os.Stat(*flagFastq); err != nil {
		abort(err)
	}

	info("processing " + *flagFastq)
	err := readGzFile(*flagFastq, *flagOut, *flagMin, *flagMax)
	if err != nil {
		abort(err)
	}
}

func info(message string) {
	fmt.Println("[ok] " + message)
}

func warn(message string) {
	fmt.Println("[* ] " + message)
}

func abort(message error) {
	fmt.Println("[!!] " + message.Error())
	os.Exit(1)
}
