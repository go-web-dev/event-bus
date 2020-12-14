package services

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

func (b *Bus) SnapshotDB(output string) error {
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	logger.Info("initiating database backup")

	var err error
	if _, e := os.Stat(output); os.IsNotExist(e) {
		err = os.Mkdir(output, os.ModePerm)
	}
	if err != nil {
		logger.Error("could not create output directory", zap.Error(err))
		return err
	}

	fileName := fmt.Sprintf("%s/backup-%s", output, time.Now().UTC().Format(time.RFC3339))
	backupFile, err := os.Create(fileName)
	if err != nil {
		logger.Error("could not create backup file", zap.Error(err))
		return err
	}

	// since: version 1 means the very first element of db
	// version increments on every update
	// the file output is a protobuf encoded list
	// you can decode it using db.Load()

	// add custom writer that writes to connection as well
	n, err := b.db.Backup(backupFile, 1)
	if err != nil {
		logger.Error("could not backup database", zap.Error(err))
		return err
	}
	if n == 0 {
		logger.Debug("nothing to backup")
	}
	logger.Info("database backup finished")
	return nil
}
