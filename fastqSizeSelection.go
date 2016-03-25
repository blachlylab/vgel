package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
)

var wg sync.WaitGroup;

type FastQrecord struct {
	header string
	sequence string
	bonus string
	quality string
}

func readGzFile(filename string, inputlines chan <- FastQrecord) {
	// read the gzipped input file
	fi, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	/*
	fz, err := gzip.NewReader(fi)
	if err != nil {
		panic(err)
	}
	defer fz.Close()
	*/

	scanner := bufio.NewScanner(fi)
	// Set up a fixed buffer to avoid allocations
	scanbuf := make([]byte, 4096);
	scanner.Buffer(scanbuf, 4096);

	for scanner.Scan() {
		fqrecord := FastQrecord{"", "", "", ""}
		fqrecord.header = scanner.Text()
		
		scanner.Scan()
		fqrecord.sequence = scanner.Text()
		
		scanner.Scan()
		fqrecord.bonus = scanner.Text()
		
		scanner.Scan()
		fqrecord.quality = scanner.Text()
		
		inputlines <- fqrecord
	}
	close(inputlines)
}

func checkLines(minLen int, maxLen int, inputlines <- chan FastQrecord, passedlines chan <- FastQrecord) {
	for fqrecord := range inputlines {
		if len(fqrecord.sequence) >= minLen && len(fqrecord.sequence) <= maxLen {
			passedlines <- fqrecord
		}
	}
	close(passedlines)
}

func writeGzFile(outFile string, passedlines <- chan FastQrecord) {
	// init writing the gzipped output file
	fo, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	/*
	fzo := gzip.NewWriter(fo)
	if err != nil {
		panic(err)
	}
	defer fzo.Close()
	*/

	for fqrecord := range passedlines {
		fo.Write( []byte(fqrecord.header + "\n") );
		fo.Write( []byte(fqrecord.sequence + "\n") );
		fo.Write( []byte(fqrecord.bonus + "\n") );
		fo.Write( []byte(fqrecord.sequence + "\n") );
	}

	// release the Wait() in main()
	wg.Done()
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
	
	inputlines := make(chan FastQrecord, 16000)
	passedlines := make(chan FastQrecord, 16000)

	info("processing " + *flagFastq)
	
	wg.Add(1) // will wait for writeGzFile to call done
	go readGzFile(*flagFastq, inputlines);
	go checkLines(*flagMin, *flagMax, inputlines, passedlines);
	go writeGzFile(*flagOut, passedlines);
	
	wg.Wait()

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
