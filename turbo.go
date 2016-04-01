package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

type FastQrecord struct {
	header   []byte
	sequence []byte
	bonus    []byte
	quality  []byte
}

func main() {
	flagFastq := flag.String("fastq", "", "fastq.gz sequence file")
	flagOut := flag.String("out", "", "output file name")
	flagMax := flag.Int("max", 1000, "max read length size")
	flagMin := flag.Int("min", 0, "min read length size")
	flag.Parse()
	// maybe do arg checking someday
	minLen := *flagMin
	maxLen := *flagMax

	var err error

	// read the input file
	var fi *os.File 
	if *flagFastq != "" {
		info("processing " + *flagFastq)
		if _, err = os.Stat(*flagFastq); err != nil {
			panic(err)
		}
		fi, err = os.Open(*flagFastq)
		if err != nil {
			panic(err)
		}
	} else {
		//read from stdin 
		info("processing STDIN")
		fi = os.Stdin
	}
	defer fi.Close()

	// init writing the output file
	var fo *os.File
	if *flagOut != "" {
		fo, err = os.Create(*flagOut)
		if err != nil {
			panic(err)
		} else {
			info("writing to " + *flagOut)
		}
	} else {
		info("writing to STDOUT")
		fo = os.Stdout
	}
	defer fo.Close()

	scanner := bufio.NewScanner(fi)
	// Set up a fixed buffer to avoid allocations
	//scanbuf := make([]byte, 16384);
	//scanner.Buffer(scanbuf, 16384);

	writer := bufio.NewWriterSize(fo, 65536)

	// Set up slices in fqrecord
	fqrecord := FastQrecord{}
	fqrecord.header = make([]byte, 1024)
	fqrecord.sequence = make([]byte, 1024)
	fqrecord.bonus = make([]byte, 1024)
	fqrecord.quality = make([]byte, 1024)

	var headlen, seqlen, bonuslen, quallen int

	newline := []byte("\n")

	for scanner.Scan() {
		// first Scan() already done
		headlen = copy(fqrecord.header, scanner.Bytes())

		scanner.Scan()
		seqlen = copy(fqrecord.sequence, scanner.Bytes())

		scanner.Scan()
		bonuslen = copy(fqrecord.bonus, scanner.Bytes())

		scanner.Scan()
		quallen = copy(fqrecord.quality, scanner.Bytes())

		if seqlen >= minLen && seqlen <= maxLen {
			writer.Write(fqrecord.header[0:headlen])
			writer.Write(newline)
			writer.Write(fqrecord.sequence[0:seqlen])
			writer.Write(newline)
			writer.Write(fqrecord.bonus[0:bonuslen])
			writer.Write(newline)
			writer.Write(fqrecord.quality[0:quallen])
			writer.Write(newline)
		}

	}
	writer.Flush()

}

func info(message string) {
	fmt.Fprintln(os.Stderr, "[ok] " + message)
}

func warn(message string) {
	fmt.Fprintln(os.Stderr, "[* ] " + message)
}

func abort(message error) {
	fmt.Fprintln(os.Stderr, "[!!] " + message.Error())
	os.Exit(1)
}
