# remotebackup

A backup automation tool built around restic.

## Motivation

In the past, I had written some bash scripts to backup some remote filesystems using a combination of dump (ext2/3/4) and GNU tar.
This had worked well enough but was not terribly extensible; I had limited support to the ext-family and btrfs, adding
XFS or bcachefs was going to take a lot of work, perhaps more than a rewrite in Go.

Just like my bash scripts, the Go reimplementation uses filesystem/volume level snapshots as a backup source for data
consistancy purposes.

