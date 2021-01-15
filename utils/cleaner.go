package utils

import (
	"io/ioutil"
	"log"
	"time"
)

func CleanOldFiles(dir string, dur time.Duration) {
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err.Error())
	}
	now := time.Now()
	for _, info := range fileInfo {
		if diff := now.Sub(info.ModTime()); diff > dur {
			log.Printf("Deleting %s which is %s old\n", info.Name(), diff)
		}
	}
}
