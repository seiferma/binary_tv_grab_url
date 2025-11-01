package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/seiferma/tvheadend/xmltv_url/internal"
	"github.com/spf13/pflag"
)

type ParsedArgs struct {
	Description  bool
	Capabilities bool
	Quiet        bool
	Output       string
	Days         int
	Offset       int
	URLs         []string
}

func main() {

	parsedArgs := parseArgs()

	error := validateArgs(parsedArgs)
	if error != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", error)
		pflag.Usage()
		return
	}

	logic := internal.GetLogic()
	error = executeBusinessLogic(parsedArgs, logic)
	if error != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", error)
		pflag.Usage()
		return
	}

}

func parseArgs() ParsedArgs {
	// arguments according to https://web.archive.org/https://wiki.xmltv.org/index.php/XmltvCapabilities
	description := pflag.Bool("description", false, "Prints name of program")
	capabilities := pflag.Bool("capabilities", false, "Prints the capabilities of program")
	quiet := pflag.Bool("quiet", false, "Suppress all output except for errors")
	output := pflag.String("output", "", "Output file (default: stdout)")
	days := pflag.Int("days", 0, "Number of days to fetch (default: infinite)")
	offset := pflag.Int("offset", 0, "Number of days to offset the start date (default: 0)")
	pflag.String("config-file", "", "not implemented - no effect")
	pflag.CommandLine.SetOutput(os.Stderr)
	pflag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage of %s [flags] <url> [<url>...]\n",
			os.Args[0],
		)
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  url (positional, at least one required)\tThe URLs to process")
	}
	pflag.Parse()

	return ParsedArgs{
		Description:  *description,
		Capabilities: *capabilities,
		Quiet:        *quiet,
		Output:       *output,
		Days:         *days,
		Offset:       *offset,
		URLs:         pflag.Args(),
	}
}

func validateArgs(parsedArgs ParsedArgs) error {
	if parsedArgs.Description && parsedArgs.Capabilities {
		return fmt.Errorf("cannot use --description and --capabilities together")
	}

	if (len(parsedArgs.URLs) < 1 || strings.TrimSpace(parsedArgs.URLs[0]) == "") && !parsedArgs.Description && !parsedArgs.Capabilities {
		return fmt.Errorf("at least one URL must be provided unless --description or --capabilities is used")
	}

	return nil
}

func executeBusinessLogic(parsedArgs ParsedArgs, logic internal.Logic) error {

	if parsedArgs.Description {
		fmt.Println(logic.GetDescriptionFunc())
		return nil
	}

	if parsedArgs.Capabilities {
		for _, cap := range logic.GetCapabilitiesFunc() {
			fmt.Println(cap)
		}
		return nil
	}

	if len(parsedArgs.URLs) < 1 {
		return fmt.Errorf("error: missing required positional argument 'url'")
	}

	request := internal.CreateRequest(
		parsedArgs.URLs,
		parsedArgs.Days,
		parsedArgs.Offset,
		parsedArgs.Quiet,
	)
	content, err := logic.GetContentFunc(request)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	if parsedArgs.Output != "" {
		err := os.WriteFile(parsedArgs.Output, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	} else {
		fmt.Println(content)
	}
	return nil
}
