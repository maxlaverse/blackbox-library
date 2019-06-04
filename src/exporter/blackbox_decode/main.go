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
	"github.com/maxlaverse/blackbox-library/src/blackbox/stream"
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("You need to provide the path to the logs")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return export(args[0], opts)
		},
	}

	cmd.Flags().IntVarP(&opts.verbose, "verbose", "v", 0, "Be verbose on log output")
	cmd.Flags().BoolVarP(&opts.raw, "raw", "", false, "Don't apply predictions to fields (show raw field deltas)")
	cmd.Flags().BoolVarP(&opts.debug, "debug", "", false, "Show extra debugging information")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected resut: %v\n", err)
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
	frameChan, err := flightLog.LoadFile(ctx, logFile)
	if err != nil {
		return err
	}

	// prepare exporter and write CSV headers
	csvExporter := exporter.NewCsvFrameExporter(bufferedWriter, opts.debug, flightLog.FrameDef)
	err = csvExporter.WriteHeaders()
	if err != nil {
		return err
	}

	for frame := range frameChan {
		// handle frame error
		if err := frame.Error(); err != nil {
			//TODO: Log offset and last id
			if !isErrorRecoverable(err) {
				cancel()
				return err
			}
		}

		// write CSV row
		err = csvExporter.WriteFrame(frame)
		if err != nil {
			cancel()
			return err
		}
	}

	return nil
}

func isErrorRecoverable(err error) bool {
	switch err.(type) {
	case *stream.ReadError, stream.ReadError:
		return false
	default:
		return true
	}
}
