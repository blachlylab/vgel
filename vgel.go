package main

// native imports
import (
	"bufio"
	"errors"
	"log"
	"os"
	"strconv"
)

// external imports
import (
	"github.com/aybabtme/uniplot/barchart"
	"github.com/codegangsta/cli"
)

// FastQrecord is a struct to represent a single fastq entry
// It contains a byte slice of each piece of the fastq entry, plus the length of each entry
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
	app.Version = "0.6.0"
	app.Usage = "Virtual Gel"
	app.Writer = os.Stderr
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
			Value: 999, // max of seqLenArray
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "keep",
			Usage:  "Extract specific sequences for analysis",
			Action: vgel,
		},
		{
			Name:  "discard",
			Usage: "Excise and discard specific sequences",
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
			Name:    "examine",
			Aliases: []string{"ex", "histogram", "hist"},
			Usage:   "Display histogram of fragment lengths",
			Action:  vgel,
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

	// ok for input and output to both be left blank (stdin/stdout)
	if input == output && input != "" {
		err := errors.New("input and output filenames shouldn't be the same")
		fatal(err, c)
	}

	//info("Mode set", c)
	info("Will consider sequences in ["+strconv.Itoa(minLen)+", "+strconv.Itoa(maxLen)+"] nt", c)

	// read the input file
	var fi *os.File
	if input != "" {
		info("reading from "+input, c)
		if _, err = os.Stat(input); err != nil {
			fatal(err, c)
		}
		fi, err = os.Open(input)
		if err != nil {
			fatal(err, c)
		}
	} else {
		//read from stdin
		info("reading from stdin", c)
		fi = os.Stdin
	}
	defer fi.Close()

	// init writing the output file
	var fo *os.File
	if output != "" {
		fo, err = os.Create(output)
		if err != nil {
			fatal(err, c)
		} else {
			info("writing to "+output, c)
		}
	} else {
		info("writing to stdout", c)
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
	// probably a substantial speed boost to use an array,
	// but limits the upper bound of fragment length, AND
	// for safety will require an if fragLen > ARRAYMAX
	// which may  mitigate some of the speed increase. Need testing.
	var seqLenArray [1000]int

	for scanner.Scan() {
		// first Scan() already done
		fqrecord.headlen = copy(fqrecord.header, scanner.Bytes())

		scanner.Scan()
		fqrecord.seqlen = copy(fqrecord.sequence, scanner.Bytes())
		if true {
			seqLenMap[fqrecord.seqlen]++
			seqLenArray[fqrecord.seqlen]++
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
		case "keep":
			if fqrecord.seqlen >= minLen && fqrecord.seqlen <= maxLen {
				writeFQrecord(&fqrecord, writer)
			}
		case "discard":
			if fqrecord.seqlen < minLen || fqrecord.seqlen > maxLen {
				writeFQrecord(&fqrecord, writer)
			}
		case "examine":
			// don't write anything
			// I am undecided whether should behave as extract, or ignore min/max
		default: // this should never happen
		}
	}
	writer.Flush()
	if c.Command.Name == "examine" {
		info("printing barchart", c)
		writeBarchart(seqLenArray, c)
	}

}

func writeBarchart(seqLenArray [1000]int, c *cli.Context) {
	var start, end int
	// Step 1. Find first and last nonzero entries in the array
	// 1a. scan forwards
	for k, v := range seqLenArray {
		start = k
		if v > 0 {
			break
		}
	}

	// 1b. scan backwards
	for i := len(seqLenArray) - 1; i >= 0; i-- {
		end = i
		if seqLenArray[i] > 0 {
			break
		}
	}

	// Step 2. Make slice of [2]int arrays
	// length = (end - start) + 1 (e.g. 9-0 + 1 = 10)
	if start == end {
		start--
	}
	data := make([][2]int, (end-start)+1)

	// Step 3. Populate the [2]int arrays from seqLenArray
	i := 0
	for j := start; j <= end; j++ {
		data[i][0] = j
		data[i][1] = seqLenArray[j]
		i++
	}
	plot := barchart.BarChartXYs(data)
	if err := barchart.Fprint(os.Stderr, plot, barchart.Linear(65)); err != nil {
		fatal(err, c)
	}

}

func info(message string, c *cli.Context) {
	log.Println("[" + c.Command.Name + "] " + message)
}

func fatal(err error, c *cli.Context) {
	log.Fatalln("[" + c.Command.Name + "][☠ ] " + err.Error())
}
