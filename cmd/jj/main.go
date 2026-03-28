package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"jumblejuice/internal/common"
	"jumblejuice/internal/encoder"
	"jumblejuice/internal/emitter"
)

var Version = "0.3.0"

func main() {
	input := flag.String("i", "", "Input file (use - for stdin)")
	enc := flag.String("e", "ipv4", "Encoder: hex, b64, dec, ipv4, ipv6, mac")
	lang := flag.String("t", "c", "Language: c, csharp, go, python, rust, nim")
	output := flag.String("o", "", "Output file (default: stdout)")
	raw := flag.Bool("raw", false, "Output data blob only, no decoder or imports")
	help := flag.Bool("h", false, "Show help")
	version := flag.Bool("v", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "JumbleJuice - Binary Payload Encoder\n")
		fmt.Fprintf(os.Stderr, "Usage: jj -i <input> [-e encoder] [-t language] [-o file] [--raw]\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEncoders:  hex, b64, dec, ipv4, ipv6, mac\n")
		fmt.Fprintf(os.Stderr, "Languages: %s\n", strings.Join(listEmitters(), ", "))
	}

	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stderr, "JumbleJuice %s\n", Version)
		os.Exit(0)
	}
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: -i is required\n")
		flag.Usage()
		os.Exit(1)
	}

	code, err := run(*input, *enc, *lang, *raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, []byte(code), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Written to %s\n", *output)
	} else {
		fmt.Print(code)
	}
}

func run(inputFile, encoderName, langName string, raw bool) (string, error) {
	data, err := common.ReadInput(inputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	enc, err := encoder.GetEncoder(encoderName)
	if err != nil {
		available := encoder.ListEncoders()
		sort.Strings(available)
		return "", fmt.Errorf("unknown encoder %q (available: %s)", encoderName, strings.Join(available, ", "))
	}

	encoded, err := enc.Encode(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	emit, err := emitter.GetEmitter(langName)
	if err != nil {
		return "", fmt.Errorf("unknown language %q (available: %s)", langName, strings.Join(listEmitters(), ", "))
	}

	return emit.Emit(encoded, raw)
}

func listEmitters() []string {
	names := emitter.ListEmitters()
	sort.Strings(names)
	return names
}
