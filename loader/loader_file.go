package loader

import (
	"fmt"
	"strconv"
	"time"
)

// Load mysql status output from a mysqladmin output file
type FileLoader struct {
	statusFile      *FileParser
	variablesFile   *FileParser
	variablesSample *Sample

	// The first uptime reported in the status file
	firstUptime int64
}

func NewFileLoader(statusFile, varFile string) *FileLoader {
	l := &FileLoader{}

	l.statusFile = NewFileParser(statusFile)
	if varFile != "" {
		l.variablesFile = NewFileParser(varFile)
	}

	return l
}

func (l *FileLoader) Initialize(interval time.Duration, sources []SourceName) error {
	// Initialize the status file loader, this has to work
	err := l.statusFile.Initialize(interval)
	if err != nil {
		return fmt.Errorf("error inititalizing status file loader: %v", err)
	}

	if l.variablesFile != nil {
		// Now initialize the variables file loader if it is set
		err = l.variablesFile.Initialize(interval)

		if err != nil {
			return fmt.Errorf("error inititalizing error file loader: %v", err)
		}

		// Currently, only a single sample on a variables file is parsed.  This is less than ideal if variables were changed over a long collection run.  Also, a potential variables file with many samples will be parsed completely, which is inefficient if we are using just one.
		l.variablesSample = l.variablesFile.GetNextSample()
		if l.variablesSample != nil && l.variablesSample.Error() != nil {
			return fmt.Errorf("error parsing variables: %v", l.variablesSample.Error())
		}

	}

	return nil
}

// Create and feed a channel of MyqSamples based on the given status and var file.
func (l *FileLoader) GetStateChannel() <-chan StateReader {
	ch := make(chan StateReader)

	sfl := l.statusFile

	// Goroutine to get status data and feed it to ch
	go func() {
		var prev_ssp *SampleSet
		for {
			// Get the next data from the Status file
			sd := sfl.GetNextSample()

			// Nil status data == EOF
			if sd == nil {
				close(ch)
				break
			}

			// Construct the new State
			state := NewState()
			state.GetCurrentWriter().SetSample(`status`, sd)
			if l.variablesSample != nil {
				// Resuse variapes sample (assume it hasn't changed)
				state.GetCurrentWriter().SetSample(`variables`, l.variablesSample)
			}
			state.SetPrevious(prev_ssp)

			// The state's uptime comes from our status file data
			if _, ok := sd.Data[`uptime`]; ok {
				// Set the uptime if we have it
				currUptime, _ := strconv.ParseInt(sd.Data[`uptime`], 10, 64)
				state.GetCurrentWriter().SetUptime(currUptime - l.firstUptime)

				// Set the first up time if we don't have it
				if l.firstUptime == 0 {
					l.firstUptime = state.GetCurrent().GetUptime()
				}
			}

			ch <- state
			prev_ssp = state.Current
		}
	}()

	return ch
}
