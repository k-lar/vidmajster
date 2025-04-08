package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

var userAgents = []string{
	// Chromium (Google Chrome)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",

	// Mozilla Firefox
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",

	// WebKit (Safari)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Safari/605.1.15",
}

func main() {
	pageURL := flag.String("url", "", "URL of the webpage to scrape for videos")
	uFlag := flag.String("u", "", "Shorthand for -url")
	iFlag := flag.String("i", "", "Alternative shorthand for -url")
	flag.Parse()

	if *pageURL == "" && *uFlag == "" && *iFlag == "" {
		fmt.Println("Please provide a URL using the -url, -u, or -i flag.")
		return
	}

	// Use the first non-empty flag value
	if *pageURL == "" {
		if *uFlag != "" {
			pageURL = uFlag
		} else {
			pageURL = iFlag
		}
	}

	videoLinks, err := scrapeWithUserAgent(*pageURL)
	if err != nil {
		fmt.Println("Scraping error:", err)
		return
	}

	if len(videoLinks) == 0 {
		fmt.Println("No video links found.")
		return
	}

	// Show found links
	fmt.Println("\nFound video links:")
	for i, link := range videoLinks {
		fmt.Printf("%d: %s\n", i+1, link)
	}

	// Prompt user for selection
	fmt.Print("\nEnter the number(s) of the video(s) to download (comma separated): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	selections := strings.Split(input, ",")

	for _, s := range selections {
		s = strings.TrimSpace(s)
		i := parseIndex(s, len(videoLinks))
		if i >= 0 {
			downloadFile(videoLinks[i])
		}
	}
}

// Use different agents for weird website logic for when firefox
// or webkit works but chrome doesn't
func scrapeWithUserAgent(targetURL string) ([]string, error) {
	parsedBase, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	var finalLinks []string
	seen := make(map[string]bool)

	for _, ua := range userAgents {
		videoLinks := []string{}

		c := colly.NewCollector(
			colly.AllowedDomains(parsedBase.Host),
			colly.MaxDepth(1),
		)

		// Spoof User-Agent
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("User-Agent", ua)
			r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
			r.Headers.Set("Referer", r.URL.String())
		})

		videoExtRegex := regexp.MustCompile(`(?i)\.(mp4|webm|ogg|mov|mkv|avi)(\?.*)?$`)

		c.OnHTML("video source", func(e *colly.HTMLElement) {
			src := resolveURL(parsedBase, e.Attr("src"))
			if !seen[src] {
				videoLinks = append(videoLinks, src)
				seen[src] = true
			}
		})

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			href := e.Attr("href")
			if videoExtRegex.MatchString(href) {
				src := resolveURL(parsedBase, href)
				if !seen[src] {
					videoLinks = append(videoLinks, src)
					seen[src] = true
				}
			}
		})

		c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
			processJSONLD(e.Text, parsedBase, seen, &videoLinks)
		})

		c.OnHTML("script", func(e *colly.HTMLElement) {
			processEmbeddedVideoInJS(e.Text, seen, &videoLinks)
		})

		err := c.Visit(targetURL)
		if err != nil {
			return nil, err
		}

		if len(videoLinks) > 0 {
			finalLinks = videoLinks
			break
		}
	}

	return finalLinks, nil
}

// Handles structured JSON-LD parsing
func processJSONLD(rawJSON string, base *url.URL, seen map[string]bool, links *[]string) {
	var data map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(rawJSON))
	dec.UseNumber()
	if err := dec.Decode(&data); err != nil {
		return
	}

	// Look for contentUrl or nested video objects
	extractVideoURLsFromJSON(data, base, seen, links)
}

// Optional: extract raw URLs from inline <script> JS blocks
func processEmbeddedVideoInJS(jsText string, seen map[string]bool, links *[]string) {
	re := regexp.MustCompile(`https?://[^\s'"<>]+?\.(mp4|webm|ogg|mov|mkv|avi)(\?[^\s'"<>]*)?`)
	found := re.FindAllString(jsText, -1)
	for _, f := range found {
		if !seen[f] {
			*links = append(*links, f)
			seen[f] = true
		}
	}
}

// Recursively search for video URLs in parsed JSON
func extractVideoURLsFromJSON(data interface{}, base *url.URL, seen map[string]bool, links *[]string) {
	switch val := data.(type) {
	case map[string]interface{}:
		for k, v := range val {
			if k == "contentUrl" || k == "url" {
				if str, ok := v.(string); ok {
					full := resolveURL(base, str)
					if !seen[full] {
						*links = append(*links, full)
						seen[full] = true
					}
				}
			} else {
				extractVideoURLsFromJSON(v, base, seen, links)
			}
		}
	case []interface{}:
		for _, item := range val {
			extractVideoURLsFromJSON(item, base, seen, links)
		}
	}
}

func resolveURL(base *url.URL, ref string) string {
	uri, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(uri).String()
}

func parseIndex(s string, max int) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil || i < 1 || i > max {
		fmt.Printf("Invalid selection: %s\n", s)
		return -1
	}
	return i - 1
}

func downloadFile(videoURL string) {
	fmt.Printf("Downloading: %s\n", videoURL)
	resp, err := http.Get(videoURL)
	if err != nil {
		fmt.Println("Download error:", err)
		return
	}
	defer resp.Body.Close()

	parsedURL, _ := url.Parse(videoURL)
	fileName := path.Base(parsedURL.Path)
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Printf("Downloaded to %s\n", fileName)
}
