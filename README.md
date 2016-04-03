# vgel: Virtual Gel
`vgel` can extract and excise sequences within specific size ranges

## Usage
```
COMMANDS:
  Alter sequences:
    extract	Extract specific sequences for analysis
    excise	Excise and discard specific sequences

  Examine sequences:
    histogram	Display histogram of fragment lengths

GLOBAL OPTIONS:
   --input, -i 		input FASTQ (default: stdin)
   --output, -o 	output FASTQ (default: stdout)
   --min, -m "0"	Minimum fragment length to consider
   --max, -M "101"	Maximum fragment length to consider
   --help, -h		show help
   --version, -v	print the version
```
