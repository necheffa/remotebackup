package remotebackup

import (
	_ "embed"
	"fmt"
	"os"

	"necheff.net/remotebackup/data"

	"github.com/bitfield/script"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed VERSION
var Version string

type RemoteBackup struct {
	conf   *data.Configuration
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

func NewRemoteBackup(conf *data.Configuration) *RemoteBackup {
	rb := new(RemoteBackup)
	rb.conf = conf

	rb.logger = func() *zap.Logger {
		encConfig := zap.NewProductionEncoderConfig()
		encConfig.TimeKey = "timestamp"
		encConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		logLevel := zap.InfoLevel
		//logLevel := zap.DebugLevel

		cfg := zap.Config{
			Level:             zap.NewAtomicLevelAt(logLevel),
			Development:       false,
			DisableCaller:     false,
			DisableStacktrace: false,
			Sampling:          nil,
			Encoding:          "json", // TODO: decide if JSON is really the encoding I want.
			EncoderConfig:     encConfig,
			OutputPaths:       []string{"stderr"}, // TODO: consider adding a log file.
			ErrorOutputPaths:  []string{"stderr"},
			InitialFields:     map[string]any{"pid": os.Getpid()},
		}

		return zap.Must(cfg.Build())
	}()

	rb.sugar = rb.logger.Sugar()

	return rb
}

func (rb *RemoteBackup) Run() {
	if err := rb.RemoteBackup(); err != nil {
		rb.sugar.Infow("Failed to run on error", "error", err)
	}
}

func (rb *RemoteBackup) RemoteBackup() error {
	for _, host := range rb.conf.Hosts {
		fmt.Println("Backing up:", host.Name)
		rb.BackupHost(&host)
	}

	return nil
}

func (rb *RemoteBackup) BackupHost(host *data.Host) error {
	snaps := []data.Snapshot{}

	// TODO: Make a factory to support other connection types based on user config.
	//       Also consider having this be a method on the Host object.
	conn := data.NewSshConnection("myhost", "myname", rb.conf)

	for _, vol := range host.Volumes {
		snap, err := data.NewSnapshot(rb.conf, conn, host, &vol)
		if err != nil {
			rb.sugar.Infow("Failed to instantiate snapshot", "error", err)
			continue
		}

		snaps = append(snaps, snap)
		err = snap.Create()
		if err != nil {
			rb.sugar.Infow("Failed to create snapshot", "error", err)
		}
	}

	for _, snap := range snaps {
		err := snap.Mount()
		if err != nil {
			rb.sugar.Infow("Failed to mount snapshot", "error", err)
			continue
		}
	}

	fmt.Println("Calling restic on:", rb.conf.Mounts+"/"+host.Name+"/")
	p := script.Exec("echo CALLING RESTIC FROM THE SHELL")
	msg, err := p.String()
	if err != nil {
		rb.sugar.Infow("Failed to execute restic", "error", err)
	} else {
		fmt.Print(msg)
	}

	for _, snap := range snaps {
		err := snap.Unmount()
		if err != nil {
			rb.sugar.Infow("Failed to unmount snapshot", "error", err)
			continue
		}

		err = snap.Destroy()
		if err != nil {
			rb.sugar.Infow("Failed to destroy snapshot", "error", err)
			continue
		}
	}

	return nil
}
