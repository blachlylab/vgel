package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
)

// import gzip "github.com/klauspost/pgzip"

var wg sync.WaitGroup
var mutex = &sync.Mutex{}

type FastQrecord struct {
	header   []byte
	sequence []byte
	bonus    []byte
	quality  []byte

	headlen  int
	seqlen   int
	bonuslen int
	quallen  int
}

func readGzFile(filename string, minLen int, maxLen int, passedlines chan<- FastQrecord) {
	// read the gzipped input file
	fi, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	/*
		fiz, err := gzip.NewReaderN(fi, 524288, 48)
		if err != nil {
			panic(err)
		}
		defer fiz.Close()
	*/
	fiz := fi
	scanner := bufio.NewScanner(fiz)

	fqrecord := FastQrecord{}
	fqrecord.header = make([]byte, 1024)
	fqrecord.sequence = make([]byte, 1024)
	fqrecord.bonus = make([]byte, 1024)
	fqrecord.quality = make([]byte, 1024)

	for scanner.Scan() {
		mutex.Lock()
		fqrecord.headlen = copy(fqrecord.header, scanner.Bytes())

		scanner.Scan()
		fqrecord.seqlen = copy(fqrecord.sequence, scanner.Bytes())

		scanner.Scan()
		fqrecord.bonuslen = copy(fqrecord.bonus, scanner.Bytes())

		scanner.Scan()
		fqrecord.quallen = copy(fqrecord.quality, scanner.Bytes())

		if fqrecord.seqlen >= minLen && fqrecord.seqlen <= maxLen {
			passedlines <- fqrecord
		} else {
			mutex.Unlock()
		}
	}
	close(passedlines)
}

//func checkLines(minLen int, maxLen int, inputlines <-chan FastQrecord, passedlines chan<- FastQrecord) {
//	for fqrecord := range inputlines {
//		if fqrecord.seqlen >= minLen && fqrecord.seqlen <= maxLen {
//			passedlines <- fqrecord
//		} else {
//			mutex.Unlock()
//		}
//	}
//	close(passedlines)
//}

func writeGzFile(outFile string, passedlines <-chan FastQrecord) {
	// init writing the gzipped output file
	fo, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	/*
		foz := gzip.NewWriter(fo)
		if err != nil {
			panic(err)
		}
		defer foz.Close()
		foz.SetConcurrency(524288, 48)
	*/
	foz := fo
	writer := bufio.NewWriterSize(foz, 65536)

	newline := []byte("\n")
	for fqrecord := range passedlines {
		writer.Write(fqrecord.header[0:fqrecord.headlen])
		writer.Write(newline)

		writer.Write(fqrecord.sequence[0:fqrecord.seqlen])
		writer.Write(newline)

		writer.Write(fqrecord.bonus[0:fqrecord.bonuslen])
		writer.Write(newline)

		writer.Write(fqrecord.quality[0:fqrecord.quallen])
		writer.Write(newline)

		mutex.Unlock()
	}
	writer.Flush()
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

	passedlines := make(chan FastQrecord, 16000)

	info("processing " + *flagFastq)

	wg.Add(1) // will wait for writeGzFile to call done
	go readGzFile(*flagFastq, *flagMin, *flagMax, passedlines)
	//go checkLines(*flagMin, *flagMax, inputlines, passedlines)
	go writeGzFile(*flagOut, passedlines)

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
