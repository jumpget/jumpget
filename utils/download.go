package utils

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-uuid"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func Download(url, downloadDir string) (fileName string, err error) {
	log.Printf("downloading file from : %v\n", url)
	splits := strings.Split(url, "/")
	fileName = splits[len(splits)-1]

	if strings.Contains(fileName, "?") {
		splits = strings.Split(fileName, "?")
		fileName = splits[0]
	}

	if fileName == "" {
		fileName, _ = uuid.GenerateUUID()
	}

	filePath := fmt.Sprintf("%s/%s", downloadDir, fileName)
	// Get the data
	resp, err := http.Get(url)
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

func DownloadWithProgress(downloadDir string, url string) (err error) {
	splits := strings.Split(url, "/")
	fileName := splits[len(splits)-1]
	filePath := fmt.Sprintf("%s/%s", downloadDir, fileName)

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(filePath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		out.Close()
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(filePath+".tmp", filePath); err != nil {
		return err
	}
	return nil
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer interface
// and we can pass this into io.TeeReader() which will report progress on each write cycle.
type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}
