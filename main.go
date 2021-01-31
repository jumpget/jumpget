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
	"github.com/lsgrep/jumpget/ssh"
	"github.com/lsgrep/jumpget/utils"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net"
	"os/user"
	"strings"
)

var (
	cfgFile     string
	sshPrivKey  string
	resourceUrl string
	sshUsername string
	host        string

	// version info
	Version   string
	BuildDate string
	GitHash   string

	sshPort     int
	server      bool
	verbose     bool
	showVersion bool
)

func init() {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	sshUsername = currentUser.Username
	hdir := currentUser.HomeDir
	if err != nil {
		panic(err)
	}
	sshPrivKey = fmt.Sprintf("%s/.ssh/id_rsa", hdir)
	cfgFile = fmt.Sprintf("%s/.jumpget.yaml", hdir)

	flag.StringVar(&sshUsername, "user", sshUsername, "ssh username")
	flag.StringVar(&host, "host", "", "jumpget server host")
	flag.StringVar(&cfgFile, "config", cfgFile, "jumpget config file")
	flag.StringVar(&sshPrivKey, "ssh-config", sshPrivKey, "ssh private key")
	flag.IntVar(&sshPort, "ssh-port", 22, "ssh port")
	flag.BoolVar(&server, "server", false, "server mode (default false)")
	flag.BoolVar(&showVersion, "version", false, "show version")
}

func prepareConfig() {
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
	flag.Parse()
}

func checkClientArgs() (err error) {
	args := flag.Args()
	if len(args) == 0 {
		return errors.New("Please pass a url to start downloading")
	}
	resourceUrl = args[0]
	if host == "" {
		cfgHost := viper.GetString("host")
		if cfgHost == "" {
			return errors.New("You have to provide a JumpGet host")
		} else {
			host = cfgHost
		}
	}

	localPort := viper.GetInt("JUMPGET_LOCAL_PORT")
	if localPort == 0 {
		viper.Set("JUMPGET_LOCAL_PORT", 4100)
	}
	return nil
}
func getIps(remote *ssh.RemoteExecutor) (ips []string, err error) {
	out, err := remote.Execute("echo $SSH_CONNECTION")
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(strings.Fields(string(out))[0])
	if ip != nil {
		ips = append(ips, ip.String())
	}

	return ips, nil
}

func main() {
	prepareConfig()
	var err error
	if showVersion {
		fmt.Printf("jumpget version: %v\nbuild date: %v\ngit version: %v\n", Version, BuildDate, GitHash)
		return
	}

	// server
	if server {
		goto serverModule
	}
	goto clientModule

serverModule:
	startServers()
	return

clientExit:
	fmt.Println(err.Error())
	return

clientModule:
	err = checkClientArgs()
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		return
	}
	remote := ssh.NewRemoteExecutor(sshPrivKey, sshUsername, host, sshPort)
	err = remote.Connect()
	if err != nil {
		goto clientExit
	}
	defer remote.Close()
	ips, err := getIps(remote)
	if err != nil {
		goto clientExit
	}
	params := struct {
		Ips []string `json:"ips"`
		Url string   `json:"url"`
	}{Ips: ips, Url: resourceUrl}

	data, err := json.Marshal(params)
	if err != nil {
		goto clientExit
	}

	// TODO check if the port is open
	command := `curl -s -H "Content-Type: application/json" -X POST --data '%s'  localhost:%d/download`
	c := fmt.Sprintf(command, string(data), viper.GetInt("JUMPGET_LOCAL_PORT"))
	// submit task
	fmt.Printf("Starting download task on server: %v\n", host)
	//fmt.Printf("command: %v\n", c)
	result, err := remote.Execute(c)
	if err != nil {
		goto clientExit
	}
	output := strings.TrimSpace(string(result))
	splits := strings.Split(output, "\n")
	newUrl := splits[len(splits)-1]
	if !utils.IsValidURL(newUrl) {
		errTemplate := `Invalid response(%v) has been returned from the JumpGet server. 
1. Check if the resource is available for download.
2. Check if JumpGet server is running at port: %v.
3. Check if JUMPGET_PUBLIC_URL is configured correctly(http or https schemes should be present).`
		err = errors.New(fmt.Sprintf(errTemplate, newUrl, viper.GetInt("JUMPGET_LOCAL_PORT")))
		goto clientExit
	}

	fmt.Printf("New location: %v, whitelisted ips: %v\n", newUrl, strings.Join(ips, ","))
	err = utils.DownloadWithProgress(".", newUrl)
	if err != nil {
		goto clientExit
	}
	return
}
