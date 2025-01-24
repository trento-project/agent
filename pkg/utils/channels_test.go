package utils_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/trento-project/agent/pkg/utils"
)

func TestCombineChannelsOnResult(t *testing.T) {
	resultChan := make(chan int)
	errorChan := make(chan error)

	go func() {
		resultChan <- 42
	}()

	result, err := utils.CombineChannels(resultChan, errorChan)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestCombineChannelsOnError(t *testing.T) {
	resultChan := make(chan int)
	errorChan := make(chan error)

	go func() {
		errorChan <- errors.New("error")
	}()

	result, err := utils.CombineChannels(resultChan, errorChan)

	assert.Error(t, err)
	assert.Zero(t, result)
}

func TestScanFileContextCancelled(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test"
	fileContent := "line1\nline2\nline3"
	err := afero.WriteFile(fs, filePath, []byte(fileContent), 0644)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	lines, err := utils.ScanFileContext(ctx, fs, filePath)

	assert.Error(t, err)
	assert.Nil(t, lines)
}

func TestScanFileContextSuccess(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test"
	fileContent := "line1\nline2\nline3"
	err := afero.WriteFile(fs, filePath, []byte(fileContent), 0644)
	assert.NoError(t, err)

	ctx := context.Background()

	lines, err := utils.ScanFileContext(ctx, fs, filePath)

	assert.NoError(t, err)
	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
}

func TestScanFileContextNoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "no-file"

	ctx := context.Background()

	lines, err := utils.ScanFileContext(ctx, fs, filePath)

	assert.Error(t, err)
	assert.Nil(t, lines)
}
