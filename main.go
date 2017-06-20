package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// Parse the command line flag values into variables for later use
	args := flagInit()

	// set up buffer characters
	var buffer []string
	if *args.bufferChars != "" {
		buffer = []string{*args.bufferChars, *args.bufferChars}
	} else if *args.bufferCharsLeft != "" || *args.bufferCharsRight != "" {
		buffer = []string{*args.bufferCharsLeft, *args.bufferCharsRight}
	} else {
		buffer = []string{"", ""}
	}

	// load tokenized document into memory
	// but first, check that the file exists
	ensureFileExists(*args.inputPath, "--input")
	input := loadFile(*args.inputPath)

	var output []byte
	if *args.extract {
		// tokens are being extracted in this block
		tokens := extractTokens(input, buffer)

		if strings.HasSuffix(*args.outputPath, ".json") {
			// if the output file ends in json, format it as a json file with { "key": "" }
			output = convertToJSON(tokens, buffer)
		} else {
			// just spit out the keys
			for i, t := range tokens {
				if i > 0 {
					output = append(output, []byte("\n")...)
				}
				output = append(output, t...)
			}
		}
	} else {
		// we are detokenizing in this block
		if *args.dataBag != "" {
			// parse tokens from data bags
			blobs := listDataBagEntries(*args.dataBag)
			if len(blobs) == 0 {
				fmt.Println("Data bag shows no entries.  Ensure you are able to view a list of data bags with the command: knife show data bags {your_databag_here}")
				os.Exit(1)
			}
			var blobsBytes [][]byte
			for _, b := range blobs {
				if *args.dataBagSecret == "" {
					blobsBytes = append(blobsBytes, collectDataBagJSON(*args.dataBag, b))
				} else {
					blobsBytes = append(blobsBytes, collectEncrytpedDataBagJSON(*args.dataBag, b, *args.dataBagSecret))
				}
			}
			var tokens []map[string][]byte
			tokens = mapKeyPairs(blobsBytes, buffer)
			output = detokenize(input, tokens)

		} else {
			paths := strings.Split(*args.tokensPath, ",")
			tokenInputs := [][]byte{}
			for _, path := range paths {
				ensureFileExists(path, "--tokens")
				tokenInputs = append(tokenInputs, loadFile(path))
			}
			// parse tokens from json file(s)
			tokens := mapKeyPairs(tokenInputs, buffer)

			// store final product for later use
			output = detokenize(input, tokens)
		}
	}

	if *args.outputPath != "" {
		outputToFile(*args.outputPath, output)
	} else {
		outputToStdout(output)
	}
}

func ensureFileExists(file, use string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Printf("Error: File \"%s\" does not exist. Please provide a valid file path for %s.\n", file, use)
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
