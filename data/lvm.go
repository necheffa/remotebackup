package data

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

type LvmSnapshot struct {
	host   *Host
	volume *Volume
	conf   *Configuration
	conn   Connection
	name   string
}

func NewLvmSnapshot(c *Configuration, conn Connection, h *Host, v *Volume) (*LvmSnapshot, error) {
	s := new(LvmSnapshot)

	s.host = h
	s.volume = v
	s.conf = c
	s.conn = conn

	return s, nil
}

func (l *LvmSnapshot) remoteMountPoint() string {
	return path.Join(l.conf.Mounts, l.volume.Name)
}

func (l *LvmSnapshot) localMountPoint() string {
	return path.Join(l.conf.Mounts, l.host.Name, l.volume.Name)
}

func (l *LvmSnapshot) Create() error {
	t := time.Now()
	n := l.volume.Name + ":rb" + fmt.Sprint(t.Unix())
	l.name = n

	createCmd := "lvcreate -y -L5G -s -pr -n " + l.name + " " + path.Join("/dev/mapper", l.volume.Name)
	if l.conf.Dryrun {
		fmt.Println(createCmd)
	}

	return nil
}

func (l *LvmSnapshot) Destroy() error {
	destroyCmd := "lvremove -y " + path.Join("/dev/mapper", l.name)
	if l.conf.Dryrun {
		fmt.Println(destroyCmd)
	}

	return nil
}

func (l *LvmSnapshot) Mount() error {
	if _, err := os.Stat(l.localMountPoint()); os.IsNotExist(err) {
		if l.conf.Dryrun {
			fmt.Println("Creating mountpoint: ", l.localMountPoint())
		} else {
			err = os.Mkdir(l.localMountPoint(), 0700)
			if err != nil && !os.IsExist(err) {
				return errors.Join(ErrUnableToCreateMountPoint, err)
			}
		}
	}

	var mountCmd string
	mountSrc := path.Join("/dev/mapper", l.name)

	switch l.volume.FileSystem {
	default:
		return ErrUnsupportedFileSystem
	case "ext4":
		mountCmd = "mount " + mountSrc + " " + l.remoteMountPoint()
	case "xfs":
		mountCmd = "mount -o ro,norecovery " + mountSrc + " " + l.remoteMountPoint()
	}

	err := l.conn.Bind(mountCmd, l.remoteMountPoint(), l.localMountPoint())
	if err != nil {
		return err
	}

	return nil
}

func (l *LvmSnapshot) Unmount() error {
	return l.conn.Unbind(l.remoteMountPoint(), l.localMountPoint())
}

func (l *LvmSnapshot) Volume() *Volume {
	return l.volume
}

func (l *LvmSnapshot) Host() *Host {
	return l.host
}
