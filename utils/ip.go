package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func GetMyIp() (ip string, err error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return ip, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ip, errors.New(fmt.Sprintf("Invalid Status Code: %v", resp.StatusCode))
	}
	data := struct {
		Ip string `json:"ip"`
	}{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	parsedIp := net.ParseIP(data.Ip)
	if parsedIp == nil {
		return "", errors.New("Invalid IP")
	}

	return parsedIp.String(), err
}

func GetRealIp(user, host string, port int, privateFile string) (ip string, err error) {
	command1 := fmt.Sprintf(`ssh -p %d %s@%s `, port, user, host)
	command2 := `"echo  \$SSH_CONNECTION"`
	result, err := RunCommand(command1 + command2)
	if err != nil {
		return
	}
	result = strings.TrimSpace(result)
	ip = strings.Fields(result)[0]
	return ip, err
}

func RunCommand(cmdStr string) (string, error) {
	c := exec.Command("sh", "-c", cmdStr)
	output := bytes.NewBufferString("")
	c.Stdout = output
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
