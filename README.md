# Eagle

Eagle is a tool for sending HTTP requests and look for patterns in the response and print if there is a match in either response body og header. 

Useful in Fingerprinting web appplications, during recon phase.

## Requirements

Golang must be installed.



## Installation

```bash
go install -v github.com/ninposec/eagle@latest
```
This will download and install the tool in your system's $GOPATH/bin directory.


Or via Git Clone:

```bash
git clone github.com/ninposec/eagle.git
cd eagle
go build .
```

### Compile binary to different platforms

Binary can easy be cross-compiled to run on Linux/Windows/MacOS.

Example:

Cross Compile to MacOS, Linux or Windows:

`GOOS=darwin GOARCH=amd64 go build`

`GOOS=windows GOARCH=amd64 go build`

`GOOS=linux GOARCH=amd64 go build`

## Usage

Eagle reads URLs from STDIN and sends HTTP requests to them. By default, it sends GET requests. You can use command-line options to customize the requests.

```bash
./eagle -h

Eagle v0.2
Author: ninposec

Find specific HTTP Responses to Fingerprint web apps.

Options:
  -b, --body <data>         Request body
  -d, --delay <delay>       Delay between issuing requests (ms)
  -H, --header <header>     Add a header to the request (can be specified multiple times)
  -hh, --hosth <string>     Insert arbitrary Host name to check for host header injection
  -up, --urlpath <string>   Insert URL Path or endpoint
  -k, --keep-alive          Use HTTP Keep-Alive
  -m, --method              HTTP method to use (default: GET, or POST if body is specified)
  -M, --match <string>      Save responses that include <string> in the body
  -fh, --findheader <string>      Show responses that include <string> in the header
  -fb, --findbody <string>      Show responses that include <string> in the body
  -o, --output <dir>        Directory to save responses in (will be created)
  -s, --save-status <code>  Save responses with given status code (can be specified multiple times)
  -S, --save                Save all responses
  -x, --proxy <proxyURL>    Use the provided HTTP proxy
  -so, --silentoutput       Do not print detailed output
  -nd, --nodebug            Suppress error messages to console
  -nr, --noredir            Suppress error messages to console
```


## Use Cases

Request URLs provided on stdin and read responses, print if hit.

Examples:

Initiate HTTP Request, look for url path and a http response that contains string (case-sensitive).

```bash
cat urls.txt | eagle -up "/endpoint" -fb "bodykeyword"
```

Initiate HTTP Request, look for specific http header and http body string in response. 

```bash
cat urls.txt | eagle -fh "headername" -fb "bodykeyword" -x http://127.0.0.1:8080 -H "x-test1: 123"
```

Look for s3 url inside the response body.

```bash
cat urls.txt | eagle -up "/" -nd -fb "s3.amazonaws.com"

https://test.explample.com/ [Found s3.amazonaws.com in HTTP Body] 
https://apps.xyz.net [Found s3.amazonaws.com in HTTP Body] 
https://assets.example.com/ [Found s3.amazonaws.com in HTTP Body] 
```

Supress errors while looking for specific file containing particular response in the body:

```
cat urls.txt | eagle -up "/package.json" -fb '"dependencies":' -nd
```

## ToDo

- switch for threadcount -c for faster run or rate limit the requests
- better error handling
- switch for timeout on requests - -timeout xx seconds
- switch to read urls from file input
- switch to read Request Headers from file input

### Other
Inspired by ProjectDiscovery´s nuclei and Tomnomnom tools.
