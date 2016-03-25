package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

type FastQrecord struct {
	header string
	sequence string
	bonus string
	quality string
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
		
		if len(fqrecord.sequence) >= minLen && len(fqrecord.sequence) <= maxLen {
			fo.Write( []byte(fqrecord.header + "\n") );
			fo.Write( []byte(fqrecord.sequence + "\n") );
			fo.Write( []byte(fqrecord.bonus + "\n") );
			fo.Write( []byte(fqrecord.quality + "\n") );

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
