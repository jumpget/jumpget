package utils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func Download(link, downloadDir string) (fileName string, err error) {
	log.Printf("downloading file from : %v\n", link)
	splits := strings.Split(link, "/")
	fileName = splits[len(splits)-1]

	if strings.Contains(fileName, "?") {
		splits = strings.Split(fileName, "?")
		fileName = splits[0]
	}

	if fileName == "" {
		fileName = uuid.New().String()
	} else {
		//
		fileName, err = url.PathUnescape(fileName)
		if err != nil {
			return "", err
		}
	}

	filePath := fmt.Sprintf("%s/%s", downloadDir, fileName)
	// Get the data
	resp, err := http.Get(link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fileName, errors.New(fmt.Sprintf("Invalid Status Code: %v", resp.StatusCode))
	}
	// Create the file
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		return fileName, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	if err = os.Rename(filePath+".tmp", filePath); err != nil {
		return
	}
	return fileName, err
}

func DownloadWithProgress(downloadDir string, link string) (err error) {
	splits := strings.Split(link, "/")
	fileName := splits[len(splits)-1]
	fileName, err = url.PathUnescape(fileName)
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("%s/%s", downloadDir, fileName)
	start := time.Now()
	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(link)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	contentLength := resp.Header.Get("Content-Length")
	total, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return err
	}
	// Create our progress reporter and pass it to be used alongside our writer
	counter := NewWriteCounter(total)
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}
	fmt.Println()
	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filePath+".tmp", filePath); err != nil {
		return err
	}

	duration := time.Since(start)
	speed := float64(total) / 1024.0 / duration.Seconds()
	fmt.Printf("Download finished in %.2f seconds, average speed: %.2f KB/s\n", duration.Seconds(), speed)
	return nil
}

// writeCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type writeCounter struct {
	Current int64
	bar     *progressBar
}

func NewWriteCounter(total int64) *writeCounter {
	counter := &writeCounter{bar: NewProgressBar(0, total, "=")}
	return counter
}

func (counter *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	counter.Current += int64(n)
	counter.PrintProgress()
	return n, nil
}

func (counter writeCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	//fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	//fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
	counter.bar.Update(counter.Current)
}
