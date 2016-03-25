package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"sync"
)

var wg sync.WaitGroup;

func readGzFile(filename string, inputlines chan <- string) {
	// read the gzipped input file
	fi, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fz, err := gzip.NewReader(fi)
	if err != nil {
		panic(err)
	}
	defer fz.Close()


	scanner := bufio.NewScanner(fz)
	// Set up a fixed buffer to avoid allocations
	scanbuf := make([]byte, 4096);
	scanner.Buffer(scanbuf, 4096);

	for scanner.Scan() {
		inputlines <- scanner.Text()
	}
	close(inputlines)
}

func checkLines(minLen int, maxLen int, inputlines <- chan string, passedlines chan <- string) {
	for fqline := range inputlines {
		if len(fqline) >= minLen && len(fqline) <= maxLen {
			passedlines <- fqline
		}
	}
	close(passedlines)
}

func writeGzFile(outFile string, passedlines <- chan string) {
	// init writing the gzipped output file
	fo, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	fzo := gzip.NewWriter(fo)
	if err != nil {
		panic(err)
	}
	defer fzo.Close()
	
	for fqline := range passedlines {
		fzo.Write( []byte(fqline + "\n") );
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
	
	inputlines := make(chan string, 400)
	passedlines := make(chan string, 400)

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
