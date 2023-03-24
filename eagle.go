package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	flag.Usage = func() {
		h := []string{
			"Eagle v0.1 ",
			"Author: ninposec ",
			"",
			"Find specific HTTP Responses to Fingerprint web apps. ",
			"",
			"Options:",
			"  -b, --body <data>         Request body",
			"  -d, --delay <delay>       Delay between issuing requests (ms)",
			"  -H, --header <header>     Add a header to the request (can be specified multiple times)",
			"  -hh, --hosth <string>     Insert arbitrary Host name to check for host header injection",
			"  -up, --urlpath <string>   Insert URL Path or endpoint",
			//"      --ignore-html         Don't save HTML files; useful when looking non-HTML files only",
			//"      --ignore-empty        Don't save empty files",
			"  -k, --keep-alive          Use HTTP Keep-Alive",
			"  -m, --method              HTTP method to use (default: GET, or POST if body is specified)",
			"  -M, --match <string>      Save responses that include <string> in the body",
			"  -fh, --findheader <string>      Show responses that include <string> in the header",
			"  -fb, --findbody <string>      Show responses that include <string> in the body",
			"  -o, --output <dir>        Directory to save responses in (will be created)",
			"  -s, --save-status <code>  Save responses with given status code (can be specified multiple times)",
			"  -S, --save                Save all responses",
			"  -x, --proxy <proxyURL>    Use the provided HTTP proxy",
			"  -so, --silentoutput       Do not print detailed output",
			"  -nd, --nodebug            Suppress error messages to console",
			"",
		}

		fmt.Fprintf(os.Stderr, strings.Join(h, "\n"))
	}
}

func main() {

	var requestBody string
	flag.StringVar(&requestBody, "body", "", "")
	flag.StringVar(&requestBody, "b", "", "")

	var keepAlives bool
	flag.BoolVar(&keepAlives, "keep-alive", false, "")
	flag.BoolVar(&keepAlives, "keep-alives", false, "")
	flag.BoolVar(&keepAlives, "k", false, "")

	var saveResponses bool
	flag.BoolVar(&saveResponses, "save", false, "")
	flag.BoolVar(&saveResponses, "S", false, "")

	var delayMs int
	flag.IntVar(&delayMs, "delay", 100, "")
	flag.IntVar(&delayMs, "d", 100, "")

	var headers headerArgs
	flag.Var(&headers, "header", "")
	flag.Var(&headers, "H", "")

	var hosth string
	flag.StringVar(&hosth, "hosth", "", "")
	flag.StringVar(&hosth, "hh", "", "")
	//flag.Var(&headers, "host", "")

	var urlpath string
	flag.StringVar(&urlpath, "urlpath", "", "")
	flag.StringVar(&urlpath, "up", "", "")
	//flag.Var(&headers, "host", "")

	var method string
	flag.StringVar(&method, "method", "GET", "")
	flag.StringVar(&method, "m", "GET", "")

	var match string
	flag.StringVar(&match, "match", "", "")
	flag.StringVar(&match, "M", "", "")

	var findheader string
	flag.StringVar(&findheader, "findheader", "", "")
	flag.StringVar(&findheader, "fh", "", "")

	var findbody string
	flag.StringVar(&findbody, "findbody", "", "")
	flag.StringVar(&findbody, "fb", "", "")

	var outputDir string
	flag.StringVar(&outputDir, "output", "out", "")
	flag.StringVar(&outputDir, "o", "out", "")

	var saveStatus saveStatusArgs
	flag.Var(&saveStatus, "save-status", "")
	flag.Var(&saveStatus, "s", "")

	var proxy string
	flag.StringVar(&proxy, "proxy", "", "")
	flag.StringVar(&proxy, "x", "", "")

	var silentOutput bool
	flag.BoolVar(&silentOutput, "silentoutput", false, "")
	flag.BoolVar(&silentOutput, "so", false, "")

	var noDebug bool
	flag.BoolVar(&noDebug, "nodebug", false, "")
	flag.BoolVar(&noDebug, "nd", false, "")

	flag.Parse()

	// if no args are provided, print flags
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	if noDebug {
		os.Stderr, _ = os.Open(os.DevNull)
	}
	

	delay := time.Duration(delayMs * 1000000)
	client := newClient(keepAlives, proxy)
	prefix := outputDir

	var wg sync.WaitGroup

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {

		rawURL := sc.Text()
		wg.Add(1)
		time.Sleep(delay)

		go func() {
			defer wg.Done()

			// create the request
			var b io.Reader
			if requestBody != "" {
				b = strings.NewReader(requestBody)

				// Can't send a body with a GET request
				if method == "GET" {
					method = "POST"
				}
			}

			_, err := url.ParseRequestURI(rawURL)
			if err != nil {
				return
			}

			req, err := http.NewRequest(method, rawURL, b)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create request: %s\n", err)
				return
			}

			if hosth != "" {
				req.Host = hosth
			}

			if urlpath != "" {
				req.URL.Path = urlpath
			}

			req.Header = map[string][]string{
				//"Content-Type":         {bodyType},
				"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36"},
				//"X-Test2":     {"test12"},
				//"Connection":    {"close"},
				//"X-FluidDB-Client-URL": {clientURL},
				//"X-FluidDB-Version":    {version},
				//"Authorization":        {"Basic " + encodedUsernameAndPassword(user, pwd)},
			}

			// add headers to the request
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)

				if len(parts) != 2 {
					continue
				}
				req.Header.Set(parts[0], parts[1])
			}

			// send the request
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "request failed: %s\n", err)
				return
			}
			defer resp.Body.Close()

			// we want to read the body into a string or something like that so we can provide options to
			// not save content based on a pattern or something like that
			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read body: %s\n", err)
				return
			}

			shouldSave := saveResponses || len(saveStatus) > 0 && saveStatus.Includes(resp.StatusCode)

			// if a -M/--match option has been used, we always want to save if it matches
			if match != "" {
				if bytes.Contains(responseBody, []byte(match)) {
					shouldSave = true
				}
			}

			statusOK := resp.StatusCode == 200

			// if a -fh/--findheader option has been used, we want to print to screen
			if findheader != "" {
				b, err := httputil.DumpResponse(resp, false)
				if err != nil {
					log.Fatalln(err)
				}
				if silentOutput == false {
					if strings.Contains(string(b), findheader) {
						fmt.Print(rawURL, urlpath, " [Found ", findheader, " in HTTP Header] ", "\n")
						//shouldSave = true
					}
				}
				// if so flag set then Only print url if there is a match in the response
				if silentOutput == true {
					if strings.Contains(string(b), findheader) {
						fmt.Print(rawURL, urlpath, "\n")
					}
				}
			}

			// if a -fb/--findbody option has been used, we want to print to screen
			if findbody != "" {
				if statusOK {
					if silentOutput == false {
						if bytes.Contains(responseBody, []byte(findbody)) {
							fmt.Print(rawURL, urlpath, " [Found ", findbody, " in HTTP Body] ", "\n")
							//shouldSave = true
						}
					}
					if silentOutput == true {
						if bytes.Contains(responseBody, []byte(findbody)) {
							fmt.Print(rawURL, urlpath, "\n")
							//shouldSave = true
						}
					}
				}
			}

			if !shouldSave {
				// Remarked due to only printing of found header and body
				//fmt.Printf("%s %d\n", rawURL, resp.StatusCode)
				return
			}
			// output files are stored in prefix/domain/normalisedpath/hash.(body|headers)
			normalisedPath := normalisePath(req.URL)
			hash := sha1.Sum([]byte(method + rawURL + requestBody + headers.String()))
			p := path.Join(prefix, req.URL.Hostname(), normalisedPath, fmt.Sprintf("%x.body", hash))
			err = os.MkdirAll(path.Dir(p), 0750)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create dir: %s\n", err)
				return
			}
			// write the response body to a file
			err = ioutil.WriteFile(p, responseBody, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to write file contents: %s\n", err)
				return
			}
			// create the headers file
			headersPath := path.Join(prefix, req.URL.Hostname(), normalisedPath, fmt.Sprintf("%x.headers", hash))
			headersFile, err := os.Create(headersPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create file: %s\n", err)
				return
			}
			defer headersFile.Close()

			var buf strings.Builder

			// put the request URL and method at the top
			buf.WriteString(fmt.Sprintf("%s %s\n\n", method, rawURL))

			// add the request headers
			for _, h := range headers {
				buf.WriteString(fmt.Sprintf("> %s\n", h))
			}
			buf.WriteRune('\n')

			// add the request body
			if requestBody != "" {
				buf.WriteString(requestBody)
				buf.WriteString("\n\n")
			}

			// add the proto and status
			buf.WriteString(fmt.Sprintf("< %s %s\n", resp.Proto, resp.Status))

			// add the response headers
			for k, vs := range resp.Header {
				for _, v := range vs {
					buf.WriteString(fmt.Sprintf("< %s: %s\n", k, v))
				}
			}

			// add the response body
			_, err = io.Copy(headersFile, strings.NewReader(buf.String()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to write file contents: %s\n", err)
				return
			}

			// output the body filename for each URL
			fmt.Printf("%s: %s %d\n", p, rawURL, resp.StatusCode)
		}()
	}
	wg.Wait()

}

func newClient(keepAlives bool, proxy string) *http.Client {

	tr := &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: !keepAlives,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 10,
			KeepAlive: time.Second,
		}).DialContext,
	}

	if proxy != "" {
		if p, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	//re := func(req *http.Request, via []*http.Request) error {
	//	return http.ErrUseLastResponse
	//}

	return &http.Client{
		Transport: tr,
		//CheckRedirect: re,
		Timeout: time.Second * 10,
	}

}

type headerArgs []string

func (h *headerArgs) Set(val string) error {
	*h = append(*h, val)
	return nil
}

func (h headerArgs) String() string {
	return strings.Join(h, ", ")
}

type saveStatusArgs []int

func (s *saveStatusArgs) Set(val string) error {
	i, _ := strconv.Atoi(val)
	*s = append(*s, i)
	return nil
}

func (s saveStatusArgs) String() string {
	return "string"
}

func (s saveStatusArgs) Includes(search int) bool {
	for _, status := range s {
		if status == search {
			return true
		}
	}
	return false
}
func normalisePath(u *url.URL) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9/._-]+`)
	return re.ReplaceAllString(u.Path, "-")
}
