package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
)

import "github.com/codegangsta/cli"

type FastQrecord struct {
	header   []byte
	sequence []byte
	bonus    []byte
	quality  []byte
}

func main() {

	var minLen, maxLen int
	var input, output string

	app := cli.NewApp()
	app.Name = "vgel"
	app.Usage = "Virtual Gel"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "input, i",
			Usage:       "input FASTQ (default: stdin)",
			Destination: &input,
		},
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "output FASTQ (default: stdout)",
			Destination: &output,
		},
		cli.IntFlag{
			Name:        "min, m",
			Usage:       "Minimum fragment length to consider",
			Destination: &minLen,
		},
		cli.IntFlag{
			Name:        "max, M",
			Usage:       "Maximum fragment length to consider",
			Destination: &maxLen,
		},
	}
	app.Commands = []cli.Command{
		{
			Category: "Alter sequences",
			Name:     "extract",
			Aliases:  []string{"ext"},
			Usage:    "Extract specific sequences for analysis",
			Action: func(c *cli.Context) {
				fmt.Println("Task: ", c.Args().First())
			},
		},
		{
			Category: "Alter sequences",
			Name:     "excise",
			Aliases:  []string{"exc"},
			Usage:    "Excise and discard specific sequences",
			Action: func(c *cli.Context) {
				fmt.Println("Task: ", c.Args().First())
			},
		},
		{
			Name:     "histogram",
			Category: "Examine sequences",
			Aliases:  []string{"hist"},
			Usage:    "Display histogram of fragment lengths",
			Action: func(c *cli.Context) {
				fmt.Println("Task: ", c.Args().First())
			},
		},
	}

	app.Action = func(c *cli.Context) {
		info("app.Action() setup ⚙ ")
		info("Minimum length: " + strconv.Itoa(minLen))
		info("Maximum length: " + strconv.Itoa(maxLen))
	}
	app.Run(os.Args)

	flagHist := flag.Bool("hist", false, "write histogram of observed read lengths")

	var err error

	// read the input file
	var fi *os.File
	if input != "" {
		info("processing " + input)
		if _, err = os.Stat(input); err != nil {
			panic(err)
		}
		fi, err = os.Open(input)
		if err != nil {
			panic(err)
		}
	} else {
		//read from stdin
		info("processing stdin")
		fi = os.Stdin
	}
	defer fi.Close()

	// init writing the output file
	var fo *os.File
	if output != "" {
		fo, err = os.Create(output)
		if err != nil {
			panic(err)
		} else {
			info("writing to " + output)
		}
	} else {
		info("writing to stdout")
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
	seqLenMap := make(map[int]int)

	for scanner.Scan() {
		// first Scan() already done
		headlen = copy(fqrecord.header, scanner.Bytes())

		scanner.Scan()
		seqlen = copy(fqrecord.sequence, scanner.Bytes())
		if *flagHist == true {
			seqLenMap[seqlen] += 1
		}

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
	if *flagHist == true {
		info("printing histogram")
		writeHist(seqLenMap)
	}
}

func writeHist(seqLenMap map[int]int) {
	keys := make([]int, len(seqLenMap))
	i := 0
	for k := range seqLenMap {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	fmt.Fprintln(os.Stderr, "\n")
	fmt.Fprintln(os.Stderr, "len", "count")
	for _, seqLen := range keys {
		fmt.Fprintln(os.Stderr, seqLen, seqLenMap[seqLen])
	}
	fmt.Fprintln(os.Stderr, "\n")
}

func info(message string) {
	fmt.Fprintln(os.Stderr, "[ok] "+message)
}

func warn(message string) {
	fmt.Fprintln(os.Stderr, "[* ] "+message)
}

func abort(message error) {
	fmt.Fprintln(os.Stderr, "[☠ ] "+message.Error())
	os.Exit(1)
}
