package skatedeluxe

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TheAldriguzDev/aldriguz_aio/packages/log"
	proxyManager "github.com/TheAldriguzDev/aldriguz_aio/packages/proxy"
	filereader "github.com/TheAldriguzDev/aldriguz_aio/packages/reader"

	tls_client "github.com/bogdanfinn/tls-client"
)

var location string = "SkateDeluxe"

func InitialiseClient(proxie string) (skateDeluxeTask, error) {
	// Add the option to do not use proxies
	proxie = strings.ReplaceAll(proxie, "\r", "")
	proxyData := strings.Split(strings.TrimSpace(proxie), ":")
	proxyUrl := fmt.Sprintf("http://%s:%s@%s:%s", proxyData[2], proxyData[3], proxyData[0], proxyData[1])

	jarOptions := []tls_client.CookieJarOption{}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithProxyUrl(proxyUrl),
		//tls_client.WithInsecureSkipVerify(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if err != nil {
		log.Error(err.Error(), location)
		return skateDeluxeTask{}, err
	}

	return skateDeluxeTask{
		httpClient: client,
	}, nil
}

func (skate *skateDeluxeTask) RotateProxy() error {
	proxie, err := proxyManager.RandomProxy(skate.ProxyArr)
	if err != nil {
		return err
	}
	proxie = strings.ReplaceAll(proxie, "\r", "")
	proxyData := strings.Split(strings.TrimSpace(proxie), ":")
	if len(proxyData) < 4 {
		return fmt.Errorf("The proxy format is not correct! %s", proxie)
	}
	proxyUrl := fmt.Sprintf("http://%s:%s@%s:%s", proxyData[2], proxyData[3], proxyData[0], proxyData[1])
	if err != nil {
		return err
	} else {
		err = skate.httpClient.SetProxy(proxyUrl)
		if err != nil {
			return err
		}
		skate.Proxy = proxie
	}
	return nil
}

func InitSkate(taskFile string, successWh string, failedWh string) {
	proxyFile, err := filereader.ReadFolder("proxies", "proxies")

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	proxyArr, err := proxyManager.ReadProxyFile(proxyFile)
	fmt.Println(proxyArr)

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	file, err := os.Open("./tasks/" + taskFile)

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	// We check if the csv file is empty, and in negative case we remove the first line that contains the table informations
	if len(rows) > 0 {
		// We remove the first line of the csv table
		rows = rows[1:]
	} else {
		log.Error("The csv file is empty!", location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	// We create a WaitGroup to syncronize the gorouting
	//var wg sync.WaitGroup
}


