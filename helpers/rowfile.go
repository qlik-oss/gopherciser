package helpers

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type (
	// RowFile path to file on disc to read into memory
	RowFile struct {
		filepath string
		readFile *sync.Once
		rows     []string
	}
)

func (file RowFile) TreatAs() string {
	return "string"
}

// NewRowFile reads rows from file path into memory
func NewRowFile(filepath string) (RowFile, error) {
	rf := RowFile{
		filepath: filepath,
		readFile: &sync.Once{},
		rows:     nil,
	}

	if err := rf.readRows(); err != nil {
		return rf, errors.WithStack(err)
	}
	return rf, nil
}

// MarshalJSON marshal filepath to JSON
func (file RowFile) MarshalJSON() ([]byte, error) {
	str := file.String()
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON reads file from filepath into memory
func (file *RowFile) UnmarshalJSON(arg []byte) error {
	file.filepath = strings.Trim(string(arg), `"`)
	if file.filepath == "" {
		return nil
	}
	file.readFile = &sync.Once{}
	if err := file.readRows(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// IsEmpty reports true if a filepath is not set
func (file *RowFile) IsEmpty() bool {
	if file == nil || file.filepath == "" {
		return true
	}
	return false
}

// String implements stringer interface
func (file RowFile) String() string {
	return file.filepath
}

// ReadRows read file into memory, will only be done once, even with multiple calls
func (file *RowFile) readRows() error {
	if runtime.GOOS == "js" {
		return nil
	}

	if file == nil || file.filepath == "" {
		return errors.New("no filepath")
	}

	var fileReadErr error
	file.readFile.Do(func() {
		bin, err := os.Open(file.filepath)
		if err != nil {
			fileReadErr = errors.Wrapf(err, "error reading from file<%s>", file.filepath)
		}
		defer func() {
			if err := bin.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing file<%s> err: %v\n", file.filepath, err)
			}
		}()

		scanner := bufio.NewScanner(bin)
		for scanner.Scan() {
			file.rows = append(file.rows, scanner.Text())
		}
		if scanner.Err() != nil {
			fileReadErr = errors.Wrapf(err, "failed reading file<%s>", file.filepath)
		}
	})
	if fileReadErr != nil {
		return fileReadErr
	}

	return nil
}

// Rows in memory from file, to read into memory, either ReadRows or UnmarshalJSON needs to have been called
func (file *RowFile) Rows() []string {
	return file.rows
}

// PurgeRows from memory
func (file *RowFile) PurgeRows() {
	file.rows = nil
}
