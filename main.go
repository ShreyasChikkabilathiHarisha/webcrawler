package main

import (
  "fmt"
  "golang.org/x/net/html"
  "io/ioutil"
  "math"
  "net"
  "net/http"
  "os"
  "strconv"
  "strings"
  "sync"
  "time"
)

type logger struct {
  mutex sync.Mutex
}

type httpClient struct {
  mutex sync.Mutex
}

type safeMap struct {
  mutex sync.Mutex
  m map[string]struct{}
}

func main(){
  if len(os.Args) < 2 {
    fmt.Println("No initial URL provided to start crawling.")
    fmt.Println("Please add the initial URL as command line argument while running the crawler.")
    return
  }

  // extract the initial URL which is the starting point of the crawl
  initURL := os.Args[1]
  fmt.Println("Starting crawling from the initial URL: ", initURL)
  fmt.Println("")

  // extract the stopping point of max number of URLs to be crawled
  maxURLCrawls := math.MaxInt32
  if len(os.Args) > 2 {
    s := os.Args[2]
    n, err := strconv.Atoi(s)
    if err == nil && n > 0 {
      maxURLCrawls = n
    }
  }

  // Initiate required structs

  // Option 3: Currently commented the map used for checking if the url has already been crawled to
  // avoid loops. I might be missing something here, but there seems to be concurrent read
  // write error on the sync map implementation, which is supposed to be thread safe. So I am using
  // own version of thread safe map with mutex which is definitely slow. We can improve the performance
  // by using something better to keep track of crawled urls

  // urlsCrawled := sync.Map{}

  log := &logger{
    mutex: sync.Mutex{},
  }
  httpClient := &httpClient{
    mutex: sync.Mutex{},
  }

  urlsAlreadyCrawled := &safeMap{
    mutex: sync.Mutex{},
    m: make(map[string]struct{}),
  }

  StartCrawlingURL(initURL, log, httpClient, urlsAlreadyCrawled, maxURLCrawls)

  return
}

// StartCrawlingURL is the starting point of web crawling for the initial url provided by the user
func StartCrawlingURL(initURL string, log *logger, httpClient *httpClient, urlsAlreadyCrawled *safeMap, maxURLCrawls int) {
  urls := ReadAndExtractURLs(initURL, httpClient)
  urlsAlreadyCrawled.m[initURL] = struct{}{}
  log.SafePrint(initURL, urls)
  CrawlURLs(urls, log, httpClient, urlsAlreadyCrawled, maxURLCrawls)
  return
}

// CrawlURLs is the thread safe recursive URL extractor for the given set of urls
func CrawlURLs(urls map[string]struct{}, log *logger, httpClient *httpClient,
  urlsAlreadyCrawled *safeMap, maxURLCrawls int) {
  // Option 2: Uncomment this to keep the crawl organized in terms of page wise crawling
  //urlMaps := make([]map[string]struct{}, 0)
  for url, _ := range urls {
    urlsAlreadyCrawled.mutex.Lock()
    _, ok := urlsAlreadyCrawled.m[url]
    urlsAlreadyCrawled.mutex.Unlock()
    if ok {
     continue
    }

    crawledURLs := ReadAndExtractURLs(url, httpClient)

    urlsAlreadyCrawled.mutex.Lock()
    urlsAlreadyCrawled.m[url] = struct{}{}

    // Check if the stopping condition has been met
    if len(urlsAlreadyCrawled.m) > maxURLCrawls {
      os.Exit(0)
    }

    urlsAlreadyCrawled.mutex.Unlock()

    log.SafePrint(url, crawledURLs)

    // Option 1: Comment this to keep the crawl organized in terms of page wise crawling
    go CrawlURLs(crawledURLs, log, httpClient, urlsAlreadyCrawled, maxURLCrawls)

    // Option 2:
    //urlMaps = append(urlMaps, crawledURLs)
  }

  // Option 2:
  //for _, m := range urlMaps {
  //  CrawlURLs(m, urlsCrawled)
  //}
}

// ReadAndExtractURLs is used to read and extract all the urls on the url specified
func ReadAndExtractURLs(initURL string, httpClient *httpClient) map[string]struct{} {
  resp, err := httpClient.GetHTTP(initURL)
  if err != nil {
    //log the error
    return map[string]struct{}{}
  }
  defer resp.Body.Close()
  htmlContent, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    //log the error
    return map[string]struct{}{}
  }

  // Parse URLs on the webpage. Using map here to avoid duplicates on the same page
  links := make(map[string]struct{})

  doc, err := html.Parse(strings.NewReader(string(htmlContent)))
  if err != nil {
    //log the error
    return map[string]struct{}{}
  }

  ExtractLinksFromHtmlNode(doc, links)

  if _, ok := links[initURL]; ok {
    delete(links, initURL)
  }

  return links
}

// ExtractLinksFromHtmlNode extracts the collects all the http and https links on the html page with href labelg
func ExtractLinksFromHtmlNode(n *html.Node, links map[string]struct{}) {
  if n.Type == html.ElementNode && n.Data == "a" {
    for _, a := range n.Attr {
      if a.Key == "href" && len(a.Val) > 6 && strings.Contains(a.Val[:6], "http") {
        l := a.Val
        strings.Trim(l, " ")
        if _, ok := links[l]; !ok {
          links[l] = struct{}{}
        }
        break
      }
    }
  }
  for c := n.FirstChild; c != nil; c = c.NextSibling {
    ExtractLinksFromHtmlNode(c, links)
  }
}

// GetHTTP defines custom transport for http client and returns the webpage response after http get
// Currently not using mutex here. Initially thought of implementing some sort of control over concurrent
// http get error due to busy port, but that doesn't seem to be required.
func (c *httpClient) GetHTTP(initURL string) (*http.Response, error) {
  //c.mutex.Lock()
  //defer c.mutex.Unlock()

  // Using custom net client to add required timeouts since the go http clients default timeout is 0 which waits forever
  var netTransport = &http.Transport{
    DialContext: (&net.Dialer{
      Timeout: 5 * time.Second,
    }).DialContext,
    TLSHandshakeTimeout: 5 * time.Second,
    IdleConnTimeout: 5 * time.Second,
    ExpectContinueTimeout: 5 * time.Second,
    MaxIdleConns: 100,
    MaxConnsPerHost: 100,
    MaxIdleConnsPerHost: 100,
  }
  var netClient = &http.Client{
    Timeout: time.Second * 10,
    Transport: netTransport,
  }

  return netClient.Get(initURL)
}

// SafePrint uses mutex to keep the print of each url and its nested urls organized in one place
// This is one of the bottlenecks and can be removed by using better loggers/ output display mechanism
func (l *logger) SafePrint(parentURL string, urls map[string]struct{}) {
  l.mutex.Lock()
  defer l.mutex.Unlock()
  fmt.Println(parentURL)
  for url, _ := range urls {
    fmt.Println("\t", url)
  }
  return
}