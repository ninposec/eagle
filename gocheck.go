package main

import (
    "fmt"
    "crypto/tls"
    //"io/ioutil"
    "os"
    "bufio"
    "time"
    "sync"
    "bytes"
    "net/http"
    //"net/http/httputil"
    "net/url"
    "net"
    "flag"
)

var concurrencyPtr = flag.Int("t", 4, "Number of threads to utilise. Default is 4.")
    //urlPtr := flag.String("url", "https://url.com","URL to Check")
var findPtr = flag.String("find", "test", "Find string in http reponse")
var proxPtr = flag.String("proxy", "test", "Proxy address - http://127.0.0.1:8080")

func init() {
    flag.Parse()
}

func main() {
    //concurrencyPtr := flag.Int("t", 4, "Number of threads to utilise. Default is 4.")
    //urlPtr := flag.String("url", "https://url.com","URL to Check")
    //findPtr := flag.String("find", "test", "Find string in http reponse")
    //proxPtr := flag.String("proxy", "test",
    //"Proxy address - http://127.0.0.1:8080")
    //flag.Parse()

    // Proxy setup
    p, err := url.Parse(*proxPtr)
    if err != nil {
        panic(err)
    }
    tr := &http.Transport{
        Proxy: http.ProxyURL(p),
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        Dial:                (&net.Dialer{Timeout: 0, KeepAlive: 0}).Dial,
        TLSHandshakeTimeout: 5 * time.Second,
        // Disable HTTP/2.
        TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
        }
    client := &http.Client{Transport: tr} 

    numWorkers := *concurrencyPtr
    work := make(chan string)
    go func() {
        s := bufio.NewScanner(os.Stdin)
        for s.Scan() {
            work <- s.Text()
        }
        close(work)
    }()

    wg := &sync.WaitGroup{}

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go doWork(work, client, wg)
    }
    wg.Wait()
}


func doWork(work chan string, client *http.Client, wg *sync.WaitGroup) {
    flag.Parse()
    defer wg.Done()
    for url := range work {
        
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
                fmt.Println(999, err)
                continue
        }
        req.Header.Add("Accept", "application/json")
        req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36")
        req.Header.Set("Connection", "close")

	    resp, err := client.Do(req)
        if err != nil {
                fmt.Println(999, err)
                continue
        }
        scan := bufio.NewScanner(resp.Body)
        toFind := []byte(*findPtr)

        defer resp.Body.Close()
        for scan.Scan() {
            if bytes.Contains(scan.Bytes(), toFind) {
                //fmt.Println(scan.Text())
                fmt.Println("Found", *findPtr, "at: ", url)
                return
            }
        }
        //if strings.Contains(resp.Header.Get("Location"),"queue") {
        //fmt.Println(resp.StatusCode, url)        
    }

}
