package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lsgrep/jumpget/utils"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var ips sync.Map

func createPublicServer(port int) *http.Server {
	downloadDir := viper.GetString("JUMPGET_DATA_DIR")

	m := http.NewServeMux()
	fs := http.FileServer(http.Dir(downloadDir))
	m.HandleFunc("/data/", func(writer http.ResponseWriter, request *http.Request) {
		bs, e := httputil.DumpRequest(request, true)
		if e != nil {
			log.Println(e)
		}
		log.Println(string(bs))
		ip := strings.Split(request.RemoteAddr, ":")[0]
		log.Printf("remote ip is: %v\n", ip)
		xRealIp := request.Header.Get("X-Real-Ip")
		forwardedIp := request.Header.Get("X-Forwarded-For")
		validIp := false
		for _, i := range []string{ip, forwardedIp, xRealIp} {
			if _, ok := ips.Load(i); ok {
				validIp = true
				break
			}
		}

		// limit access to the whitelisted IPs
		if validIp {
			http.StripPrefix("/data/", fs).ServeHTTP(writer, request)
		} else {
			log.Printf("invalid ip: %v", []string{ip, forwardedIp, xRealIp})
			writer.WriteHeader(http.StatusNotFound)
		}
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: m,
	}
	return &server
}

func createLocalServer(port int) *http.Server {
	downloadDir := viper.GetString("JUMPGET_DATA_DIR")

	type downloadData struct {
		Url string   `json:"url"`
		Ips []string `json:"ips"`
	}
	r := mux.NewRouter()
	r.HandleFunc("/download", func(writer http.ResponseWriter, request *http.Request) {
		params := mux.Vars(request)
		log.Printf("params: %v\n", params)
		log.Printf("body: %v\n", request.Body)
		log.Printf("remote address: %v\n", request.RemoteAddr)

		decoder := json.NewDecoder(request.Body)
		var data downloadData
		err := decoder.Decode(&data)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}

		// whitelist access
		log.Printf("adding ips: %v to the whitelist\n", data.Ips)
		for _, ip := range data.Ips {
			ips.Store(ip, true)
		}
		//downloadUrl := utils.FromBase64(data.Url)
		fileName, err := utils.Download(data.Url, downloadDir)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		filePath := fmt.Sprintf("%v/%v", downloadDir, fileName)
		fi, err := os.Stat(filePath)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
		// get the size
		size := fi.Size()
		writer.Write([]byte(fmt.Sprintf("download completed on the server, file size: %v\n", size)))
		publicUrl := viper.GetString("JUMPGET_PUBLIC_URL")
		writer.Write([]byte(fmt.Sprintf("%s/data/%s\n", publicUrl, url.PathEscape(fileName))))

	}).Methods(http.MethodPost)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}
	return &server
}

func startServers() {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		publicPort := viper.GetInt("JUMPGET_PUBLIC_PORT")
		pubServer := createPublicServer(publicPort)
		log.Printf("starting public server :%v\n", publicPort)
		log.Println(pubServer.ListenAndServe())
		wg.Done()
	}()

	go func() {
		localPort := viper.GetInt("JUMPGET_LOCAL_PORT")
		localServer := createLocalServer(localPort)
		log.Printf("starting local server :%v\n", localPort)
		log.Println(localServer.ListenAndServe())
		wg.Done()
	}()

	go func() {
		downloadDir := viper.GetString("JUMPGET_DATA_DIR")
		for {
			h := viper.GetInt64("JUMPGET_FILE_RETAIN_DURATION")
			if h == 0 {
				h = 12
			}
			dur := h * int64(time.Hour)
			utils.CleanOldFiles(downloadDir, time.Duration(dur))
			time.Sleep(time.Hour)
		}

	}()
	wg.Wait()
}
