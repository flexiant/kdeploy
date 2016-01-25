package utils

import (
	log "github.com/Sirupsen/logrus"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
