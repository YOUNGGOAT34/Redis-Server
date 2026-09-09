package aof

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AOF struct {
	Dir            string
	AppendOnly     string
	AppendDirName  string
	AppendFilename string
	AppendFsync    string
	Sequence       int
	File           *os.File
}

func (aofConfig *AOF) CreateAOFDir() error {

	if aofConfig.AppendOnly == "no" {
		return nil
	}

	aofDir := filepath.Join(aofConfig.Dir, aofConfig.AppendDirName)

	if err := os.MkdirAll(aofDir, 0755); err != nil {
		return err
	}

	manifestPath := filepath.Join(aofDir, BuildManifestFileName(aofConfig.AppendFilename))

	if _, err := os.Stat(manifestPath); err == nil {
		filename, err := ReadManifest(manifestPath)

		if err != nil {
			return err
		}
		aofPath := filepath.Join(aofDir, filename)

		file, err := os.OpenFile(aofPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

		if err != nil {
			return err
		}

		aofConfig.File = file
		return nil

	}

	aofConfig.Sequence = 1

	aofFileName := aofConfig.buildAOFFileName(aofConfig.AppendFilename)

	aofPath := filepath.Join(aofDir, aofFileName)

	aofFile, err := os.OpenFile(aofPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return err
	}

	aofConfig.File = aofFile

	return os.WriteFile(manifestPath, fmt.Appendf(nil, "file %s type i sequence %d\n", aofFileName, aofConfig.Sequence), 0644)
}

// AppendCommand writes a command's raw bytes to the AOF file and applies the
// configured --appendfsync durability policy:
//
//   - "always": fsync immediately after this write, so the command is durable
//     on disk before returning (slowest, strongest guarantee).
//   - "everysec": no fsync here - durability is handled by a background
//     ticker (see StartFsyncLoop) that fsyncs at most once per second,
//     matching Redis's tradeoff without a queueing/batching system.
//   - "no" (or any other/unset value): never explicitly fsync; the write is
//     left buffered in the OS page cache until the OS flushes it on its own.
//
// If no AOF file is open (appendonly disabled), this is a no-op.
func (aofConfig *AOF) AppendCommand(commandBytes []byte) error {
	if aofConfig.File == nil {
		return nil
	}

	if _, err := aofConfig.File.Write(commandBytes); err != nil {
		return err
	}

	if aofConfig.AppendFsync == "always" {
		return aofConfig.File.Sync()
	}

	return nil
}

// StartFsyncLoop launches a background goroutine that fsyncs the AOF file
// once per second, for as long as appendonly is enabled and appendfsync is
// "everysec". It is a no-op for any other appendfsync value. The goroutine
// exits when stop is closed, and closes done right before returning, so a
// caller that wants to guarantee the goroutine has fully stopped touching
// aofConfig.File (e.g. before calling aofConfig.File.Close() on shutdown)
// can wait on done rather than relying on stop's close alone - closing a
// channel only wakes the receiver, it doesn't wait for it to finish running.
// If StartFsyncLoop is a no-op, done is closed immediately so callers can
// unconditionally wait on it either way.
func (aofConfig *AOF) StartFsyncLoop(stop <-chan struct{}) (done <-chan struct{}) {
	doneCh := make(chan struct{})

	if aofConfig.AppendOnly != "yes" || aofConfig.AppendFsync != "everysec" {
		close(doneCh)
		return doneCh
	}

	go func() {
		defer close(doneCh)

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if aofConfig.File != nil {
					_ = aofConfig.File.Sync()
				}
			case <-stop:
				return
			}
		}
	}()

	return doneCh
}

func (aofConfig *AOF) buildAOFFileName(baseName string) string {
	return fmt.Sprintf("%s.%d.incr.aof", baseName, aofConfig.Sequence)
}

func BuildManifestFileName(aofFilename string) string {
	return fmt.Sprintf("%s.manifest", aofFilename)
}

func ReadManifest(manifestPath string) (string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(string(data))

	if len(parts) < 2 {
		return "", errors.New("invalid manifest")
	}

	if parts[0] != "file" {
		return "", errors.New("invalid manifest format")
	}

	return parts[1], nil
}
