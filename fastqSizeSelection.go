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
	// Set up a fixed buffer to avoid allocations
	scanbuf := make([]byte, 4096);
	scanner.Buffer(scanbuf, 4096);

	header := make([]byte, 1024);
	sequence := make([]byte, 1024);
	bonus := make([]byte, 1024);
	quality := make([]byte, 1024);
	
	var headlen, seqlen, bonuslen, quallen int;
	
	for scanner.Scan() {
		/* fastq format :
			1: header line
			2: sequence line
			3: header line (may only be "+")
			4: phred quality score line
		must parse through the file 4 lines at a time
		*/
		
		// for loop performs initial Scan()
		headlen = copy(header, scanner.Bytes());
		
		scanner.Scan()
		seqlen = copy(sequence, scanner.Bytes());
		
		scanner.Scan()
		bonuslen = copy(bonus, scanner.Bytes());
		
		scanner.Scan()
		quallen = copy(quality, scanner.Bytes());
		
		// check if this read meets our size requirements
		if seqlen >= min && seqlen <= max {
			fzo.Write(header[0:headlen])
			fzo.Write([]byte("\n"))
			
			fzo.Write(sequence[0:seqlen])
			fzo.Write([]byte("\n"))
			
			fzo.Write(bonus[0:bonuslen])
			fzo.Write([]byte("\n"))
			
			fzo.Write(quality[0:quallen])
			fzo.Write([]byte("\n"))
			/*
			for _, line := range [][]byte{header, sequence, bonus, quality} {
				//fmt.Println(line)
				fzo.Write([]byte(line + []byte("\n")))
			} */
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
