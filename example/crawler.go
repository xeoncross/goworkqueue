package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xeoncross/goworkqueue"
	"golang.org/x/net/html"
)

/*
 * Complex example of using goworkqueue to crawl a domain.
 * Based on https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html
 *
 * Run:
 *  $ go run crawler.go https://httpbin.org/links/5
 *
 * @todo: Make foundUrls a concurrent-safe map
 */

var foundUrls = make(map[string]bool)
var foundDomains = make(map[string]bool)
var queue goworkqueue.Queue
var domainBackoff *cache.Cache

func main() {

	// We want to only hit the same domain at *most* every X minutes
	domainBackoff = cache.New(1*time.Second, 2*time.Second)
	// foundUrls = make(map[string]bool)
	seedUrls := os.Args[1:]

	jobQueueSize := 1000
	numberOfWorkers := 3

	queue = goworkqueue.Queue{}
	queue.Init(jobQueueSize, numberOfWorkers, crawlWorker)

	// Abort when we press CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
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
	fmt.Println("\nFound", len(foundUrls), "unique urls:")
	for url := range foundUrls {
		fmt.Println(" - " + url)
	}

	fmt.Println("\nFound", len(foundDomains), "unique domains:")
	for url := range foundDomains {
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
func crawlWorker(url string, workerID int) {

	domain := domainOfURL(url)

	// Too soon
	if _, found := domainBackoff.Get(domain); found {
		// fmt.Println("WAIT:", domain, "->", url)
		queue.Jobs <- url
		return
	}

	// Set the value of the key "foo" to "bar", with the default expiration time
	domainBackoff.Set(domain, true, cache.DefaultExpiration)

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

			url = toAbsURL(resp.Request.URL, url)

			if _, ok := foundUrls[url]; ok {
				fmt.Println("ALREADY PARSED:", url)
				return
			}

			// Make sure the url begines in http**
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				domain := domainOfURL(url)
				foundDomains[domain] = true
				foundUrls[url] = true
				queue.Jobs <- url
			}
		}
	}
}

func toAbsURL(baseurl *url.URL, weburl string) string {
	relurl, err := url.Parse(weburl)
	if err != nil {
		return ""
	}
	absurl := baseurl.ResolveReference(relurl)
	return absurl.String()
}

func domainOfURL(weburl string) string {
	parsedURL, err := url.Parse(weburl)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}
