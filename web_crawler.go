// Refernce: https://www.hoodoo.digital/blog/how-to-make-a-web-crawler-using-go-and-colly-tutorial
package main

import (
	"fmt"
	"log"
	"github.com/gocolly/colly"
)

func main() {
	// Define the baseURL and startingURL.
	baseURL := "www.example.com"
	startingURL := "https://" + baseURL

	// Specify the allowedUrls, limited to only the baseURL.
	allowedUrls := []string{baseURL}

	// Print the startingURL for debugging purposes.
	fmt.Println(startingURL)

	// Create a new Colly collector (crawler) with the specified settings.
	c := colly.NewCollector(
		colly.AllowedDomains(allowedUrls...), // Set the allowed domains.
		colly.MaxDepth(0),                    // Set the maximum depth of the crawling process (0 means no limit).
		colly.IgnoreRobotsTxt(),              // Ignore the robots.txt file.
	)

	// The following are various callback functions to handle different events during the crawling process:

	// OnRequest is triggered when the program sends a request to the server.
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	// OnError is triggered if an error occurs while processing a request.
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	// OnResponse is triggered when the program receives a response from the server.
	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited", r.Request.URL)
	})

	// OnHTML is triggered when the program accesses an HTML resource.
	// It looks for anchor tags and then recursively visits the links they point to.
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	// OnHTML is triggered when the program accesses an HTML resource.
	// It looks for table rows (tr) and prints the content of the first column (td:nth-of-type(1)).
	c.OnHTML("tr td:nth-of-type(1)", func(e *colly.HTMLElement) {
		fmt.Println("First column of a table row:", e.Text)
	})

	// OnXML is triggered if the program receives an XML resource rather than an HTML resource.
	// In this case, it looks for h1 tags and prints their text.
	c.OnXML("//h1", func(e *colly.XMLElement) {
		fmt.Println(e.Text)
	})

	// OnScraped is triggered after the program finishes scraping a resource.
	// It prints the URL of the finished resource.
	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished", r.Request.URL)
	})

	// Start the crawling process by visiting the startingURL.
	c.Visit(startingURL)
	
}
