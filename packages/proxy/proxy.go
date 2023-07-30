package proxyManager

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

func ReadProxyFile(path string) ([]string, error) {
	content, err := ioutil.ReadFile("./proxies/" + path)

	if err != nil {
		var emptyArray []string
		return emptyArray, err
	}

	proxyList := strings.Split(string(content), "\n")
	return proxyList, nil
}

func RandomProxy(proxyList []string) (string, error) {
	if len(proxyList) == 0 {
		return "", errors.New("No proxy has been found!")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(proxyList))

	return proxyList[randomIndex], nil
}
