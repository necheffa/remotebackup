package main

import (
	"fmt"
	"log"

	"necheff.net/remotebackup"
	"necheff.net/remotebackup/data"
)

func main() {
	config, err := data.NewConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	if config.Version {
		msg := "remotebackup v" + remotebackup.Version + "\n" +
			"Copyright (C) 2020, 2024 Alexander Necheff\n" +
			"remotebackup is licensed under the terms of the GPLv3."
		fmt.Println(msg)
		return
	}

	rb := remotebackup.NewRemoteBackup(config)
	rb.Run()
}
