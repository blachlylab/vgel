package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

type FastQrecord struct {
	header []byte
	sequence []byte
	bonus []byte
	quality []byte
}


func main() {
	flagFastq := flag.String("fastq", "", "fastq.gz sequence file")
	flagOut := flag.String("out", "output.fastq.gz", "output file name")
	flagMax := flag.Int("max", 1000, "max read length size")
	flagMin := flag.Int("min", 0, "min read length size")
	flag.Parse()
	// maybe do arg checking someday
	minLen := *flagMin;
	maxLen := *flagMax;

	if _, err := os.Stat(*flagFastq); err != nil {
		abort(err)
	}
	
	info("processing " + *flagFastq)
	
	// read the input file
	fi, err := os.Open(*flagFastq)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	// init writing the output file
	fo, err := os.Create(*flagOut)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	
	scanner := bufio.NewScanner(fi)
	// Set up a fixed buffer to avoid allocations
	//scanbuf := make([]byte, 16384);
	//scanner.Buffer(scanbuf, 16384);
	
	writer := bufio.NewWriterSize(fo, 65536);

	// Set up slices in fqrecord
	fqrecord := FastQrecord{}
	fqrecord.header = make([]byte, 1024);
	fqrecord.sequence = make([]byte, 1024);
	fqrecord.bonus = make([]byte, 1024);
	fqrecord.quality = make([]byte, 1024);
	
	var headlen, seqlen, bonuslen, quallen int;

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
			writer.Write( fqrecord.header[0:headlen]  );
			writer.Write( newline );
			writer.Write( fqrecord.sequence[0:seqlen] );
			writer.Write( newline );
			writer.Write( fqrecord.bonus[0:bonuslen] );
			writer.Write( newline );
			writer.Write( fqrecord.quality[0:quallen] );
			writer.Write( newline );
		}
	
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
