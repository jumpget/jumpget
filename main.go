/*
Copyright © 2021 Yusup

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lsgrep/jumpget/utils"
	"github.com/mitchellh/go-homedir"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/user"
	"strings"
	"sync"
	"time"
)

var cfgFile string
var sshPrivateKeyFile string

var resourceUrl string
var sshUsername string
var host string
var sshPort int

var server bool

func getIps() (ips []string) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		ip1, err := utils.GetMyIp()
		if err != nil {
			panic(err)
		}
		ips = append(ips, ip1)
		wg.Done()
	}()

	go func() {
		ip2, err := utils.GetRealIp(sshUsername, host, sshPort, sshPrivateKeyFile)
		if err != nil {
			panic(err)
		}
		ips = append(ips, ip2)
		wg.Done()
	}()
	wg.Wait()
	return
}

var ips sync.Map

func createPublicServer(port int, downloadDir string) *http.Server {
	m := http.NewServeMux()
	fs := http.FileServer(http.Dir(downloadDir))
	m.HandleFunc("/data/", func(writer http.ResponseWriter, request *http.Request) {
		bs, e := httputil.DumpRequest(request, true)
		if e != nil {
			fmt.Println(e)
		}
		log.Println(string(bs))
		ip := strings.Split(request.RemoteAddr, ":")[0]
		log.Printf("remote ip is: %v\n", ip)
		// limit access
		if _, ok := ips.Load(ip); ok {
			http.StripPrefix("/data/", fs).ServeHTTP(writer, request)
		} else {
			writer.WriteHeader(http.StatusNotFound)
		}
	})

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: m,
	}
	return &server
}

func createLocalServer(port int, downloadDir string) *http.Server {
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
		writer.Write([]byte(fmt.Sprintf("%s/data/%s\n", publicUrl, fileName)))

	}).Methods(http.MethodPost)

	r.HandleFunc("/download/jobs", func(writer http.ResponseWriter, request *http.Request) {
		params := mux.Vars(request)
		id := params["id"]
		writer.Write([]byte(id))
	}).Methods(http.MethodGet)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: r,
	}
	return &server
}

func init() {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	sshUsername = currentUser.Username
	hdir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	sshPrivateKeyFile = fmt.Sprintf("%s/.ssh/id_rsa", hdir)
	cfgFile = fmt.Sprintf("%s/.jumpget.yaml", hdir)
}

func main() {
	flag.StringVar(&sshUsername, "user", sshUsername, "ssh username")
	flag.StringVar(&host, "host", "", "jumpget server host")
	flag.StringVar(&cfgFile, "config", cfgFile, "config file")
	flag.StringVar(&sshPrivateKeyFile, "ssh-config", sshPrivateKeyFile, "ssh private key")
	flag.IntVar(&sshPort, "ssh-port", 22, "ssh port")
	flag.BoolVar(&server, "server", false, "server mode (default false)")
	flag.Parse()

	// parse config file
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println(err)
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
		}
	}
	viper.AutomaticEnv()

	if !server {
		args := flag.Args()
		if len(args) == 0 {
			fmt.Println("Please pass a url to start downloading")
			flag.Usage()
			return
		}
		resourceUrl = args[0]
		fmt.Printf("Jumpget task: %v\n", resourceUrl)
		if host == "" {
			cfgHost := viper.GetString("host")
			if cfgHost == "" {
				fmt.Println("You have to provide jumpget host")
				flag.Usage()
				return
			} else {
				host = cfgHost
			}
		}
		ips := getIps()
		localPort := viper.GetInt("JUMPGET_LOCAL_PORT")
		if localPort == 0 {
			localPort = 4100
		}

		params := struct {
			Ips []string `json:"ips"`
			Url string   `json:"url"`
		}{Ips: ips, Url: resourceUrl}

		data, err := json.Marshal(params)
		if err != nil {
			panic(err)
		}

		command := `curl -s -H "Content-Type: application/json" -X POST --data '%s'  localhost:%d/download`
		c := fmt.Sprintf(command, string(data), localPort)
		fmt.Printf("Starting download task on server: %v\n", host)
		// submit task
		result := utils.SshCommand(sshPrivateKeyFile,
			sshUsername, host, sshPort, c)
		result = strings.TrimSpace(result)
		splits := strings.Split(strings.TrimSpace(result), "\n")
		newDownloadUrl := splits[len(splits)-1]
		fmt.Printf("New location: %v, whitelisted ips: %v\n", newDownloadUrl, strings.Join(ips, ","))
		err = utils.DownloadWithProgress(".", newDownloadUrl)
		if err != nil {
			panic(err)
		}
		return
	}

	// server mode
	wg := new(sync.WaitGroup)
	wg.Add(2)

	downloadDir := "/home/jumpget/data"
	go func() {
		publicPort := viper.GetInt("JUMPGET_PUBLIC_PORT")
		pubServer := createPublicServer(publicPort, downloadDir)
		log.Printf("starting public server :%v\n", publicPort)
		log.Println(pubServer.ListenAndServe())
		wg.Done()
	}()

	go func() {
		localPort := viper.GetInt("JUMPGET_LOCAL_PORT")
		localServer := createLocalServer(localPort, downloadDir)
		log.Printf("starting local server :%v\n", localPort)
		log.Println(localServer.ListenAndServe())
		wg.Done()
	}()

	go func() {
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
