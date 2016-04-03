# vgel: Virtual Gel
`vgel` can extract and excise sequences within specific size ranges

## Usage
```
COMMANDS:
  Alter sequences:
    keep	Extract specific sequences for analysis
    discard	Excise and discard specific sequences

  Examine sequences:
    examine	Display histogram of fragment lengths

GLOBAL OPTIONS:
   --input, -i 		input FASTQ (default: stdin)
   --output, -o 	output FASTQ (default: stdout)
   --min, -m "0"	Minimum fragment length to consider
   --max, -M "101"	Maximum fragment length to consider
   --help, -h		show help
   --version, -v	print the version
```
