package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/maxlaverse/blackbox-library/src/blackbox"
	"github.com/maxlaverse/blackbox-library/src/exporter/exporter"
	"github.com/spf13/cobra"
)

type cmdOptions struct {
	raw     bool
	debug   bool
	verbose int
}

func main() {
	var opts cmdOptions

	cmd := &cobra.Command{
		Use: "blackbox_decode [options] <input logs>",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.Set("logtostderr", "true")
			flag.Set("v", strconv.Itoa(opts.verbose))
			flag.CommandLine.Parse([]string{})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return export(args[0], opts)
		},
	}

	cmd.Flags().IntVarP(&opts.verbose, "verbose", "v", 0, "Be verbose on log output")
	cmd.Flags().BoolVarP(&opts.raw, "raw", "", false, "Don't apply predictions to fields (show raw field deltas)")
	cmd.Flags().BoolVarP(&opts.raw, "debug", "", false, "Show extra debugging information")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func export(sourceFilepath string, opts cmdOptions) error {
	filename := path.Base(sourceFilepath)
	dirpath := path.Dir(sourceFilepath)
	parts := strings.Split(filename, ".")
	csvFilepath := path.Join(dirpath, fmt.Sprintf("%s01.csv", strings.TrimSuffix(filename, parts[len(parts)-1])))

	logFile, err := os.Open(sourceFilepath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// prepare reader and target file
	readerOpts := blackbox.FlightLogReaderOpts{Raw: opts.raw}
	flightLog := blackbox.NewFlightLogReader(readerOpts)
	defer func() {
		fmt.Println(flightLog.Stats)
	}()

	csvFile, err := os.Create(csvFilepath)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	bufferedWriter := bufio.NewWriter(csvFile)

	// iterate over frames and write them to CSV
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	frameChan, errChan, err := flightLog.LoadFile(logFile, ctx)
	if err != nil {
		return err
	}

	// prepare exporter and write CSV headers
	csvExporter := exporter.NewCsvFrameExporter(bufferedWriter, opts.debug, flightLog.FrameDef)

	headersSent := false
	for {
		select {
		// read frames and write to CSV
		case frame := <-frameChan:
			// write headers before first row
			if !headersSent {
				err = csvExporter.WriteHeaders()
				headersSent = true
			}
			if err != nil {
				return err
			}

			// write CSV row
			err = csvExporter.WriteFrame(frame)
			if err != nil {
				cancel()
				return err
			}

		// read errors
		case err := <-errChan:
			cancel()
			return err

		default:
		}
	}
}
