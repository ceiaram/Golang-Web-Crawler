// Refernce: https://www.hoodoo.digital/blog/how-to-make-a-web-crawler-using-go-and-colly-tutorial
package main

import (
	"bufio"
	"strings"
	"os"
	"fmt"
	"log"
	"net/url"
	"github.com/gocolly/colly"
)
// CustomError is a struct to hold error messages and additional information.
type CustomError struct {
	Message string
	Code    int
}

// Implement the error interface for CustomError.
func (e CustomError) Error() string {
	return fmt.Sprintf("Error: %s (Code: %d)", e.Message, e.Code)
}

//Check URL structure and SEO-friendliness
func checkURLValidity(inputURL string) []error {
	var errors []error
	u, err := url.Parse(inputURL)

	if err != nil {
		// Return a CustomError instance for the parsing error
		errors = append(errors, CustomError{
			Message: "There was a parsing error",
			Code: 500,
		})
	}

	// Check for a valid scheme (http or https)
	if u.Scheme != "http" && u.Scheme != "https" {
		// Return a CustomError instance for the HTTP error
		errors = append(errors, CustomError{
			Message: "HTTP Error",
			Code: 400,
		})
	}

	// Check for SEO-friendliness of the URL
	// Define the maximum allowed URL length
	maxURLLength := 100

	// Compare the length of the URL's path with the maximum allowed length
	if len(u.Path) > maxURLLength {
		errors = append(errors, CustomError{
			Message: "SEO-Friendliness violation length of URL's path is over 100",
			Code: 199,
		})
	}
	
	// Check if the URL path is descriptive and readable (no query strings or fragments)
	if u.RawQuery != "" || u.Fragment != "" {
		errors = append(errors, CustomError{
			Message: "SEO-Friendliness violation URL's path of either the query string or fragment is empty",
			Code: 199,
		})
	}

	// Check if the URL path contains only lowercase letters and hyphens (SEO-friendly characters)
	if u.Path != strings.ToLower(u.Path) || strings.Contains(u.Path, "_") {
		errors = append(errors, CustomError{
			Message: "SEO-Friendliness Violation URL's path contains uppercase letters and/or hyphens",
			Code: 199,
		})
	}

	// Return all errors
	return errors
}


func main() {
	// Define the baseURL list for user input
	inputURLs := [] string {}

	// Create scanner to read user input (urls)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter website urls (separated by spaces): ")

	// Read the input from the user
	for scanner.Scan(){
		input := scanner.Text()

		// Split the input into individual URLs (assuming they are separated by spaces)
		urls := strings.Fields(input)

		// Check if valid URLs are provided 
		if len(urls) > 0 {
			//Append input to list 
			inputURLs = append(inputURLs, urls...)
			break
		} else {
			fmt.Println("Please enter at least one valid URL.")
			fmt.Print("Enter website urls (separated by spaces): ")
		}
	}

	// Print the list of URLs
	fmt.Println("List of URLs:", inputURLs)

	// Create a new Colly collector (crawler) with the specified settings.
	c := colly.NewCollector(
		colly.MaxDepth(0),                    // Set the maximum depth of the crawling process (0 means no limit).
		colly.Async(true),          // Enable asynchronous scraping.
		colly.IgnoreRobotsTxt(),              // Ignore the robots.txt file.
	)

	// The following are various callback functions to handle different events during the crawling process:
	// Set a custom User-Agent string in the request headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	})

	// OnError is triggered if an error occurs while processing a request.
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Something went wrong, https:", err)
	})

	// OnResponse is triggered when the program receives a response from the server.
	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Visited: %s (Status: %d, Content-Type: %s)\n", r.Request.URL, r.StatusCode, r.Headers.Get("Content-Type"))
	})

	// OnHTML is triggered when the program accesses an HTML resource.
	// It looks for the page title and prints it.
	c.OnHTML("title", func(e *colly.HTMLElement) {
		fmt.Printf("Page Title: %s\n", e.Text)
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

	// Start the crawling process by visiting the inputURLs and checking for url validity.
	for _, inputURL := range inputURLs {
		// Handle the list of errors returned by checkURLValidity()
		errors := checkURLValidity(inputURL)
		if len(errors) > 0 {
			fmt.Println("url: ", inputURL)
			for _, err := range errors {
				fmt.Println(err.Error())
			}
		} else {
			// If the URL is valid, visit it using colly
			c.Visit(inputURL)
		}
	}
	
	// Wait for the asynchronous scraping to finish.
	c.Wait()
}
