package utils

import (
	"bufio"
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Takes two channels and return the tuple (result, error) as soon as one of them receive data
func CombineChannels[R any](resultChan <-chan R, errorChan <-chan error) (R, error) {
	var zero R
	select {
	case err := <-errorChan:
		return zero, err
	case result := <-resultChan:
		return result, nil
	}
}

// Read a file line by line, and return the array of string for each line.
// The provided context is used to interrupt the scan if a cancellation signal occur
// On cancellation, the context error is returned
func ScanFileContext(ctx context.Context, fs afero.Fs, filePath string) ([]string, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	file, err := fs.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open "+filePath)
	}

	defer file.Close()

	errChan := make(chan error, 1)
	contentChan := make(chan []string)

	go func() {

		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		var fileLines []string

		for fileScanner.Scan() {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				fileLines = append(fileLines, fileScanner.Text())
			}
		}

		contentChan <- fileLines
	}()

	return CombineChannels(contentChan, errChan)

}
