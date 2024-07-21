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
	conn := data.NewSshConnection(host.Name, rb.conf.User, rb.conf)

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

	err := os.Setenv("RESTIC_PASSWORD_FILE", rb.conf.PasswdFile)
	if err != nil {
		rb.sugar.Infow("Failed to set restic password file environment variable", "host", host.Name, "error", err)
		// TODO: clean up, this is basically fatal.
	}

	// Use a single repository, rather than per-host, to reap the benifits of deduplication.
	// A large number of OS files are going to be deuplicated across all hosts.
	err = os.Setenv("RESTIC_REPOSITORY", rb.conf.Mounts+"/repo/")
	if err != nil {
		rb.sugar.Infow("Failed to set restic repository environment variable", "host", host.Name, "error", err)
		// TODO: cleanup, this is basically fatal.
	}

	repoPipe := script.IfExists(rb.conf.Mounts + "/repo/config")
	_, err = repoPipe.String()
	if err != nil {
		initCmd := "restic init"
		if rb.conf.Dryrun {
			fmt.Println("mkdir -p " + rb.conf.Mounts + "/repo/ && chmod 0700 " + rb.conf.Mounts + "/repo/")
			fmt.Println(initCmd)
		} else {
			err = os.MkdirAll(rb.conf.Mounts+"/repo/", 0700)
			if err != nil {
				rb.sugar.Infow("Failed to create restic repo dir", "error", err)
				// TODO: cleanup, this is basically fatal.
			}

			p := script.Exec(initCmd)
			msg, err := p.String()
			if err != nil {
				rb.sugar.Infow("Failed to initalize restic repo", "error", err)
				// TODO: cleanup, this is basically fatal.
			} else {
				fmt.Println(msg)
			}
		}
	}

	// TODO: since we are picking up /srv/remote/backup/host.Name off of sshfs, are we effectively downloading
	// a bunch of files that restic dedup will hash and determine don't need archived?
	bkupCmd := "restic backup " + rb.conf.Mounts + "/" + host.Name + "/"
	if rb.conf.Dryrun {
		fmt.Println(bkupCmd)
	} else {
		p := script.Exec(bkupCmd)
		msg, err := p.String()
		if err != nil {
			rb.sugar.Infow("Failed to execute restic backup", "host", host.Name, "error", err)
		} else {
			fmt.Println(msg)
		}
	}

	// TODO: allow the user to specify a number of days to keep backups for, default to 14.
	rotateCmd := "restic forget --host " + host.Name + " -d 14"
	if rb.conf.Dryrun {
		fmt.Println(rotateCmd)
	} else {
		p := script.Exec(rotateCmd)
		msg, err := p.String()
		if err != nil {
			rb.sugar.Infow("Failed to execute restic forget", "host", host.Name, "error", err)
		} else {
			fmt.Println(msg)
		}
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
