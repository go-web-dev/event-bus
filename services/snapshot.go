package services

import (
	"os"

	"go.uber.org/zap"

	"github.com/chill-and-code/event-bus/logging"
)

func (b *Bus) SnapshotDB(output string) error {
	// create directory and store backups with timestamps
	logger := logging.Logger
	b.mu.Lock()
	defer b.mu.Unlock()

	logger.Info("initiating database backup")
	backupFile, err := os.Create(output)
	if err != nil {
		logger.Error("could not create backup file", zap.Error(err))
		return err
	}

	// since: version 1 means the very first element of db
	// version increments on every update
	// the file output is a protobuf encoded list
	// you can decode it using db.Load()
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
