package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
)

import "github.com/codegangsta/cli"

type FastQrecord struct {
	headlen  int
	seqlen   int
	bonuslen int
	quallen  int
	header   []byte
	sequence []byte
	bonus    []byte
	quality  []byte
}

func main() {

	app := cli.NewApp()
	app.Name = "vgel"
	app.Version = "0.5.0"
	app.Usage = "Virtual Gel"
	app.Authors = []cli.Author{
		{
			Name:  "James S Blachly, MD",
			Email: "james.blachly@gmail.com",
		},
		{
			Name:  "Karl Kroll",
			Email: "kwkroll32@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "input FASTQ (default: stdin)",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "output FASTQ (default: stdout)",
		},
		cli.IntFlag{
			Name:  "min, m",
			Usage: "Minimum fragment length to consider",
		},
		cli.IntFlag{
			Name:  "max, M",
			Usage: "Maximum fragment length to consider",
			Value: 101,
		},
	}
	app.Commands = []cli.Command{
		{
			Category: "Alter sequences",
			Name:     "extract",
			Aliases:  []string{"ext"},
			Usage:    "Extract specific sequences for analysis",
			Action:   vgel,
		},
		{
			Category: "Alter sequences",
			Name:     "excise",
			Aliases:  []string{"exc"},
			Usage:    "Excise and discard specific sequences",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "save, s",
					Usage: "Save excised reads as FASTQ",
					Value: "",
				},
			},
			Action: vgel,
		},
		{
			Category: "Examine sequences",
			Name:     "histogram",
			Aliases:  []string{"hist", "histo"},
			Usage:    "Display histogram of fragment lengths",
			Action: func(c *cli.Context) {
				fmt.Println("Task: ", c.Args().First())
			},
		},
	}
	/*
		app.Action = func(c *cli.Context) {
			cli.ShowAppHelp(c)
			info("app.Action() setup ⚙ ")
			info("Will consider sequences in [" + strconv.Itoa(minLen) + ", " + strconv.Itoa(maxLen) + "] nt")
		}
	*/
	app.Action = cli.ShowAppHelp
	app.Run(os.Args)
}

func vgel(c *cli.Context) {
	var err error

	input := c.GlobalString("input")
	output := c.GlobalString("output")
	minLen := c.GlobalInt("min")
	maxLen := c.GlobalInt("max")

	if input == output {
		err := errors.New("input and output filenames shouldn't be the same")
		abort(err)
	}

	info("Mode: " + c.Command.Name)
	info("Will consider sequences in [" + strconv.Itoa(minLen) + ", " + strconv.Itoa(maxLen) + "] nt")

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

	newline := []byte("\n")
	seqLenMap := make(map[int]int)

	for scanner.Scan() {
		// first Scan() already done
		fqrecord.headlen = copy(fqrecord.header, scanner.Bytes())

		scanner.Scan()
		fqrecord.seqlen = copy(fqrecord.sequence, scanner.Bytes())
		if true {
			seqLenMap[fqrecord.seqlen] += 1
		}

		scanner.Scan()
		fqrecord.bonuslen = copy(fqrecord.bonus, scanner.Bytes())

		scanner.Scan()
		fqrecord.quallen = copy(fqrecord.quality, scanner.Bytes())

		// Write FastQ record to buffer
		writeFQrecord := func(fqr *FastQrecord, writer *bufio.Writer) {
			writer.Write(fqrecord.header[0:fqrecord.headlen])
			writer.Write(newline)
			writer.Write(fqrecord.sequence[0:fqrecord.seqlen])
			writer.Write(newline)
			writer.Write(fqrecord.bonus[0:fqrecord.bonuslen])
			writer.Write(newline)
			writer.Write(fqrecord.quality[0:fqrecord.quallen])
			writer.Write(newline)
		}
		switch c.Command.Name {
		case "extract":
			if fqrecord.seqlen >= minLen && fqrecord.seqlen <= maxLen {
				writeFQrecord(&fqrecord, writer)
			}
		case "excise":
			if fqrecord.seqlen < minLen || fqrecord.seqlen > maxLen {
				writeFQrecord(&fqrecord, writer)
			}
		default: // this should never happen
		}

	}
	writer.Flush()
	if false {
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
