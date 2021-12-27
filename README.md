# eagle

Request URLs provided on stdin and read responses, print if hit.

Example:

cat urls.txt | ./eagle -fh "X-Dr-Appname" -fb "tick" -x http://127.0.0.1:8080 -H "x-test1: 123"

Usage:


./eagle -h

```
Request URLs provided on stdin and read responses 

Options:
  -b, --body <data>         Request body
  -d, --delay <delay>       Delay between issuing requests (ms)
  -H, --header <header>     Add a header to the request (can be specified multiple times)
  -hh, --hosth <string>     Insert arbitrary Host name to check for host header injection
  -k, --keep-alive          Use HTTP Keep-Alive
  -m, --method              HTTP method to use (default: GET, or POST if body is specified)
  -M, --match <string>      Save responses that include <string> in the body
  -fh, --findheader <string>      Show responses that include <string> in the header
  -fb, --findbody <string>      Show responses that include <string> in the body
  -o, --output <dir>        Directory to save responses in (will be created)
  -s, --save-status <code>  Save responses with given status code (can be specified multiple times)
  -S, --save                Save all responses
  -x, --proxy <proxyURL>    Use the provided HTTP proxy
  
```
