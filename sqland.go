package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"math/rand"
	"time"
	"io/ioutil"
	"strings"
	"sync"
	"golang.org/x/net/html"
)

var (
	// Dork global fokus ke param id=
	dorksList = []string{
		"inurl:index.php?id=", "inurl:view.php?id=", "inurl:detail.php?id=",
		"inurl:read.php?id=", "inurl:profile.php?id=", "inurl:news.php?id=",
		"inurl:artikel.php?id=", "inurl:page.php?id=", "inurl:berita.php?id=",
		"inurl:info.php?id=", "inurl:product.php?id=", "inurl:content.php?id=",
		"inurl:show.php?id=", "inurl:more.php?id=", "inurl:info-detail.php?id=",
		"inurl:item.php?id=", "inurl:viewnews.php?id=", "inurl:viewitem.php?id=",
		"inurl:post.php?id=", "inurl:event.php?id=", "inurl:report.php?id=",
		"inurl:download.php?id=", "inurl:service.php?id=", "inurl:ticket.php?id=",
		"inurl:display.php?id=", "inurl:watch.php?id=", "inurl:record.php?id=",
		"inurl:detail_news.php?id=", "inurl:show_news.php?id=", "inurl:gallery.php?id=",
		"inurl:viewphoto.php?id=", "inurl:announce.php?id=", "inurl:ad.php?id=",
		"inurl:sport.php?id=", "inurl:doc.php?id=", "inurl:announcement.php?id=",
		"inurl:result.php?id=", "inurl:viewfile.php?id=", "inurl:event_detail.php?id=",
		"inurl:view_article.php?id=", "inurl:view.php?record_id=", "inurl:product.php?item_id=",
		"inurl:display.php?content_id=", "inurl:profile.php?user_id=", "inurl:shownews.php?id=",
		"inurl:doc_read.php?id=", "inurl:showimage.php?id=", "inurl:more_news.php?id=",
		"inurl:berita.php?id=",
	}

	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/91.0.4472.114",
		"Mozilla/5.0 (Linux; Android 10) Chrome/91.0.4472.101",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:89.0) Gecko/20100101 Firefox/89.0",
	}

	bingBaseURL = "https://www.bing.com/search?q="
	idParamRegex = regexp.MustCompile(`(^|&)id[a-zA-Z0-9_]*=`)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getHTML(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func isLinkValid(link string) bool {
	parsed, err := url.Parse(link)
	if err != nil || parsed.Host == "" || strings.Contains(parsed.Host, "bing.com") || parsed.RawQuery == "" { // Baris ini yang diperbaiki
		return false
	}
	return idParamRegex.MatchString(parsed.RawQuery)
}

func searchDork(dork string, maxPages int, minDelay time.Duration, wg *sync.WaitGroup, foundUrls chan<- string) {
	defer wg.Done()
	
	for pageNum := 0; pageNum < maxPages; pageNum++ {
		queryURL := fmt.Sprintf("%s%s&first=%d", bingBaseURL, url.QueryEscape(dork), pageNum*10+1)
		
		htmlContent, err := getHTML(queryURL)
		if err != nil {
			break
		}
		
		doc, err := html.Parse(strings.NewReader(htmlContent))
		if err != nil {
			break
		}

		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, a := range n.Attr {
					if a.Key == "href" && isLinkValid(a.Val) {
						foundUrls <- a.Val
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
		time.Sleep(minDelay)
	}
}

func printBanner() {
	fmt.Println("=== SQLand By HeruGanz1337 ===")
}

func main() {
	printBanner()
	
	var wg sync.WaitGroup
	foundUrls := make(chan string)
	uniqueUrls := make(map[string]struct{})
	
	fmt.Println("\n--- Mulai pencarian target ---\n")
	
	for _, dork := range dorksList {
		wg.Add(1)
		go searchDork(dork, 5, 5*time.Second, &wg, foundUrls)
	}

	go func() {
		wg.Wait()
		close(foundUrls)
	}()

	for url := range foundUrls {
		if _, exists := uniqueUrls[url]; !exists {
			fmt.Println(url)
			uniqueUrls[url] = struct{}{}
		}
	}
	
	fmt.Println("\n--- Selesai ---")
	fmt.Printf("Total ditemukan: %d URL\n", len(uniqueUrls))
}