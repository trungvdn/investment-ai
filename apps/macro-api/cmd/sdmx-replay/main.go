// sdmx-replay parses a captured pilot XML with the existing Go parser.
// It performs no network requests and opens no database connection.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/providers"
	"io"
	"os"
)

func main() {
	dataset := flag.String("dataset", "", "nso_cpi or nso_iip")
	file := flag.String("file", "", "captured XML path")
	flag.Parse()
	code, series, name := "VN_CPI_YOY", "PCPI_IX", "CPIVNM"
	switch *dataset {
	case "nso_cpi":
	case "nso_iip":
		code, series, name = "VN_IIP_YOY", "AIP_ISIC4_IX", "IIPVNM"
	default:
		fmt.Fprintln(os.Stderr, "unsupported pilot dataset")
		os.Exit(1)
	}
	f, err := os.Open(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "captured XML unavailable")
		os.Exit(1)
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, 16*1024*1024+1))
	if err != nil || len(body) > 16*1024*1024 {
		fmt.Fprintln(os.Stderr, "captured XML unreadable or too large")
		os.Exit(1)
	}
	rows, err := providers.ParseSDMX(body, code, series, providers.NSDPBase+name+".xml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Go parser rejected captured XML")
		os.Exit(1)
	}
	if json.NewEncoder(os.Stdout).Encode(rows) != nil {
		os.Exit(1)
	}
}
