package data

import (
	"errors"
	"fmt"

	"github.com/bitfield/script"
)

var (
	ErrSshConnection = errSshConnection()
	ErrSshLoadKeys   = errSshLoadKeys()
)

type SshConnection struct {
	host       string
	user       string
	keysLoaded bool
	conf       *Configuration
}

func NewSshConnection(host, user string, config *Configuration) *SshConnection {
	sc := new(SshConnection)
	sc.host = host
	sc.user = user
	sc.keysLoaded = false
	sc.conf = config

	return sc
}

func (sc *SshConnection) loadKeys() error {
	fmt.Println("loading SSH keys...")
	return nil
}

func (sc *SshConnection) Bind(mountCmd, remoteMount, localMount string) error {
	if !sc.keysLoaded {
		err := sc.loadKeys()
		if err != nil {
			return errors.Join(ErrSshLoadKeys, err)
		}
	}

	// TODO: Kinda gross. Doesn't really handle if remoteMount exists and is not a directory.
	createMntPoint := "if [ ! -f " + remoteMount + " ]; then mkdir -p " + remoteMount + " && chmod 0700 " + remoteMount + " fi;"

	sshCmd := "ssh " + sc.user + "@" + sc.host + " " + createMntPoint + " " + mountCmd
	if sc.conf.Dryrun {
		fmt.Println(sshCmd)
	} else {
		p := script.Exec(sshCmd)
		msg, err := p.String()
		if err != nil {
			return errors.Join(ErrSshConnection, ErrBindFailure, err)
		}
		fmt.Println(msg)
	}

	sshFsCmd := "sshfs " + sc.user + "@" + sc.host + ":" + remoteMount + " " + localMount
	if sc.conf.Dryrun {
		fmt.Println(sshFsCmd)
	} else {
		p := script.Exec(sshFsCmd)
		msg, err := p.String()
		if err != nil {
			return errors.Join(ErrSshConnection, ErrBindFailure, err)
		}
		fmt.Println(msg)
	}

	return nil
}

func (sc *SshConnection) Unbind(umountCmd, remoteMount, localMount string) error {
	if !sc.keysLoaded {
		err := sc.loadKeys()
		if err != nil {
			return errors.Join(ErrSshLoadKeys, err)
		}
	}

	localMountCmd := "umount " + localMount
	if sc.conf.Dryrun {
		fmt.Println(localMountCmd)
	} else {
		p := script.Exec(localMountCmd)
		msg, err := p.String()
		if err != nil {
			return errors.Join(ErrSshConnection, ErrUnbindFailure, err)
		}
		fmt.Println(msg)
	}

	remoteMountCmd := "ssh " + sc.user + "@" + sc.host + " " + umountCmd
	if sc.conf.Dryrun {
		fmt.Println(remoteMountCmd)
	} else {
		p := script.Exec(remoteMountCmd)
		msg, err := p.String()
		if err != nil {
			return errors.Join(ErrSshConnection, ErrUnbindFailure, err)
		}
		fmt.Println(msg)
	}

	return nil
}

func errSshConnection() error {
	return errors.New("sshconnection")
}

func errSshLoadKeys() error {
	return errors.New("loading ssh keys failed")
}
