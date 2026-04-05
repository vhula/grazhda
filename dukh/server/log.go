package server

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// maxLogSizeMiB is the maximum size of the log file before rotation.
const maxLogSizeMiB = 5

// InitLogger configures charmbracelet/log to write to $GRAZHDA_DIR/logs/dukh.log
// with 5 MiB rotation (3 backups, compressed). It also mirrors output to stderr
// so operators can tail the process directly. Returns the logger and a cleanup fn.
func InitLogger(grazhdaDir string) (*log.Logger, func(), error) {
	logDir := filepath.Join(grazhdaDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, nil, err
	}

	rotator := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "dukh.log"),
		MaxSize:    maxLogSizeMiB, // MiB
		MaxBackups: 3,
		Compress:   true,
	}

	// Write to both the rotated file and stderr (for direct process observation).
	multi := io.MultiWriter(rotator, os.Stderr)
	logger := log.New(multi)
	logger.SetLevel(log.InfoLevel)

	cleanup := func() { _ = rotator.Close() }
	return logger, cleanup, nil
}
