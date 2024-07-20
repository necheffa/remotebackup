package data

import (
	"errors"
	"fmt"

	"github.com/bitfield/script"
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
			return err
		}
	}

	// TODO: Kinda gross. Doesn't really handle if remoteMount exists and is not a directory.
	createMntPoint := "if [ ! -f " + remoteMount + " ]; then mkdir -p " + remoteMount + " && chmod 0700 " + remoteMount + " fi;"

	sshCmd := "ssh " + sc.user + "@" + sc.host + " " + createMntPoint + " " + mountCmd
	if sc.conf.Dryrun {
		fmt.Println(sshCmd)
	} else {
		p := script.Exec(sshCmd)
		if p.ExitStatus() != 0 {
			//TODO: figure out messaging
			//return errors.New("SshConnection: ", p.Stdout())
			return errors.New("sshconnection")
		}
	}

	sshFsCmd := "sshfs " + sc.user + "@" + sc.host + ":" + remoteMount + " " + localMount
	if sc.conf.Dryrun {
		fmt.Println(sshFsCmd)
	} else {
		p := script.Exec(sshFsCmd)
		if p.ExitStatus() != 0 {
			//TODO: figure out messaging
			//return errors.New("SshConnection: ", p.Stdout())
			return errors.New("sshconnection")
		}
	}

	return nil
}

func (sc *SshConnection) Unbind(remoteMount, localMount string) error {
	if !sc.keysLoaded {
		err := sc.loadKeys()
		if err != nil {
			return err
		}
	}

	localMountCmd := "umount " + localMount
	if sc.conf.Dryrun {
		fmt.Println(localMountCmd)
	} else {
		p := script.Exec(localMountCmd)
		if p.ExitStatus() != 0 {
			//TODO: figure out messaging
			//return errors.New("SshConnection: ", p.Stdout())
			return errors.New("sshconnection")
		}
	}

	remoteMountCmd := "umount " + remoteMount
	if sc.conf.Dryrun {
		fmt.Println(remoteMountCmd)
	} else {
		p := script.Exec(remoteMountCmd)
		if p.ExitStatus() != 0 {
			//TODO: figure out messaging
			//return errors.New("SshConnection: ", p.Stdout())
			return errors.New("sshconnection")
		}
	}

	return nil
}
