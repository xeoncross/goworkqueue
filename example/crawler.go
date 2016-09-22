package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
  "net/url"
	"os"
  "os/signal"
	"strings"
  "github.com/xeoncross/goworkqueue"
)

/*
 * Complex example of using goworkqueue to crawl a domain.
 * Based on https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
 *
 * Run:
 *  $ go run crawler.go https://httpbin.org/links/5
 *
 * Todo:
 *  1) Make foundUrls a concurrent-safe map
 *  2) Add https://github.com/patrickmn/go-cache to keep a domain rate-limit list
 */

var foundUrls map[string]bool
var queue goworkqueue.Queue

func main() {
  	foundUrls = make(map[string]bool)
  	seedUrls := os.Args[1:]

    jobQueueSize := 1000
    numberOfWorkers := 3

    queue = goworkqueue.Queue{}
    queue.Init(jobQueueSize, numberOfWorkers, crawlWorker)

    // Abort when we press CTRL+C
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func(){
        for _ = range c {
            queue.Close()
            fmt.Println("ABORT!")
        }
    }()

    // Add our urls to the job list
    for _, url := range seedUrls {
      queue.Jobs <- url
    }

    // Blocks until queue.Close()
    queue.Run()

    // Optional, callback for emptying the queue *if* anything remains
    queue.Drain(func(job string) {
      fmt.Printf("'%s' wasn't fetched\n", job)
    })

  	// We're done! Print the results...
  	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

  	for url, _ := range foundUrls {
  		fmt.Println(" - " + url)
  	}

}


// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Extract all http** links from a given webpage
func crawlWorker(url string, workerId int) {

  fmt.Println("fetching", url)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}

      // fmt.Println("URL:", url)
      url = toAbsUrl(resp.Request.URL, url)
      // fmt.Println("ABS URL:", url)

      if _, ok := foundUrls[url]; ok {
        fmt.Println("NO", url)
        return
      }

			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
        foundUrls[url] = true
				queue.Jobs <- url
			}
		}
	}
}

func toAbsUrl(baseurl *url.URL, weburl string) string {
	relurl, err := url.Parse(weburl)
	if err != nil {
		return ""
	}
	absurl := baseurl.ResolveReference(relurl)
	return absurl.String()
}
