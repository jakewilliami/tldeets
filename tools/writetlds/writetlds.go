package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/biter777/countries"
	"github.com/nfx/go-htmltable"

	"github.com/jakewilliami/tldinfo/pkg/tldinfo"
)

// https://stackoverflow.com/a/38644571
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	rootpath   = filepath.Dir(filepath.Dir(basepath))
)

type TLD struct {
	Domain  string          `header:"Domain"`
	Type    tldinfo.TLDType `header:"Type"`
	Manager string          `header:"TLD Manager"`
}

func checkWriterErr(err error, file string) {
	if err != nil {
		fmt.Printf("[ERROR] Could not write line to file \"%s\": %s", file, err)
		os.Exit(1)
	}
}

func isASCII(s string) bool {
	for _, char := range s {
		if char > 127 {
			return false
		}
	}
	return true
}

func main() {
	fmt.Printf("[INFO] Found base module path at %s\n", rootpath)

	htmltable.Logger = func(_ context.Context, msg string, fields ...any) {
		fmt.Printf("[INFO] %s %v\n", msg, fields)
	}

	// https://stackoverflow.com/a/74328802
	url := "https://www.iana.org/domains/root/db"
	table, err := htmltable.NewSliceFromURL[TLD](url)
	if err != nil {
		fmt.Printf("[ERROR] Could not get table by %s: %s", url, err)
		os.Exit(1)
	}

	dataRaw := make(map[string]TLD, len(table))
	for i := 0; i < len(table); i++ {
		tld := table[i]
		dataRaw[tld.Domain] = tld
	}

	data := make(map[string]tldinfo.TLD, len(dataRaw))
	for tldStr, tld := range dataRaw {
		var country string
		// TODO: this will not always work; e.g. Saint Helena is has ccTLD .ac,
		// but country code SH.  Another example: .su is for Soviet Union, but
		// as it is no longer a country (e.g., ISO 3166-3).
		// NOTE: while biter777/countries has domains, it's not complete
		if tld.Type == tldinfo.CountryCode {
			var countryCode string
			if tldStr[0] == '.' {
				countryCode = tldStr[1:]
			}
			countryCode = strings.ToUpper(countryCode)
			country = countries.ByName(countryCode).Info().Name
			if country == "Unknown" {
				country = ""
			}
		}
		data[tldStr] = tldinfo.TLD{
			Domain:  tld.Domain,
			Type:    tld.Type,
			Manager: tld.Manager,
			Country: country,
		}
	}

	writeMode := "const"
	allowedWriteModes := []string{"const", "json"}
	if len(os.Args) > 1 {
		writeMode = os.Args[1]
		if !slices.Contains(allowedWriteModes, writeMode) {
			fmt.Printf("[ERROR] Invalid write mode \"%s\"; allowed modes: %v\n", writeMode, allowedWriteModes)
			os.Exit(1)
		}
	} else {
		fmt.Printf("[ERROR] Must specify a write mode to output; allowed modes: %v\n", allowedWriteModes)
		os.Exit(1)
	}

	if writeMode == "const" {
		pkgName := "tldinfo"
		outFile := filepath.Join(rootpath, "pkg", pkgName, "tldsconst.go")

		file, err := os.Create(outFile)
		if err != nil {
			fmt.Printf("[ERROR] Cannot open file \"%s\": %s\n", outFile, err)
			os.Exit(1)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		_, err = writer.WriteString("package " + pkgName + "\n\n")
		checkWriterErr(err, outFile)
		_, err = writer.WriteString("var (\n")
		checkWriterErr(err, outFile)

		tldsSkipped := 0
		for _, tld := range data {
			if !isASCII(tld.Domain) {
				tldsSkipped++
				continue
			}

			var tldPrefix string
			if tld.Domain[0] == '.' {
				tldPrefix = tld.Domain[1:]
			}
			if tld.Type == tldinfo.CountryCode {
				tldPrefix = strings.ToUpper(tldPrefix)
			} else {
				tldPrefix = strings.Title(tldPrefix)
			}
			// TODO: is this a good naming scheme for these constants?
			_, err = writer.WriteString(fmt.Sprintf("%sTopLevelDomain = TLD{\nDomain: \"%s\",\nType: \"%s\",\nManager: %s,\nCountry: \"%s\",\n}\n", tldPrefix, tld.Domain, tld.Type, strconv.Quote(tld.Manager), tld.Country))
			checkWriterErr(err, outFile)
		}

		fmt.Printf("[WARNING] Skipped %d non-unicode domain name(s)\n", tldsSkipped)

		_, err = writer.WriteString(")\n")
		checkWriterErr(err, outFile)

		err = writer.Flush()
		if err != nil {
			fmt.Printf("[ERROR] Could not flush file writer: %s", err)
			os.Exit(1)
		}

		fileInfo, err := os.Stat(outFile)
		if err == nil {
			fmt.Printf("[INFO] Wrote %d bytes to \"%s\"\n", fileInfo.Size(), outFile)
		}

		// TODO: Would like to do this without calling to external command
		// Consider using: https://github.com/mvdan/gofumpt
		cmd := exec.Command("go", "fmt", outFile)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("[WARNING] Failed to run `go fmt` on output file \"%s\": %s\n", outFile, err)
		} else {
			fmt.Printf("[INFO] Successfully ran `go fmt` on output file \"%s\"\n", outFile)
		}
	} else if writeMode == "json" {
		tldJson, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Printf("[ERROR] Could not JSONify data: %s\n", err)
			os.Exit(1)
		}

		outFile := filepath.Join(rootpath, "assets", "tlds.json")
		err = ioutil.WriteFile(outFile, tldJson, 0644)
		if err != nil {
			fmt.Printf("[ERROR] Count not write JSON output to %s: %s", outFile, err)
			os.Exit(1)
		}

		// fmt.Printf("[DEBUG] %+v\n", data)
		fmt.Printf("[INFO] Wrote %d bytes to \"%s\"\n", len(tldJson), outFile)
	}
}
