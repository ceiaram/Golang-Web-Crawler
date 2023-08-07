// Refernce(s): https://www.hoodoo.digital/blog/how-to-make-a-web-crawler-using-go-and-colly-tutorial
// https://code.visualstudio.com/docs/cpp/config-mingw
// https://developer.fyne.io/
package main

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gocolly/colly/v2"
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


// ... (Imports and CustomError struct)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Web Crawler")
	myWindow.Resize(fyne.NewSize(800, 450))

	entry := widget.NewEntry()
	entry.SetPlaceHolder("https://www.example.com")

	// Create a crawling logs label.
	crawlingLogsLabel := widget.NewLabel("Crawling Logs")
	crawlingLogsLabel.Alignment = fyne.TextAlignLeading

	crawlingLogs := widget.NewLabel("")

	scrollContainer := container.NewVScroll(crawlingLogs)
	scrollContainer.SetMinSize(fyne.NewSize(800, 500))

	startButton := widget.NewButtonWithIcon("Start Crawling", theme.ComputerIcon(), func() {
		inputURLs := strings.Fields(entry.Text)
		if len(inputURLs) > 0 {
			crawlingLogs.SetText("") // Clear the crawling logs before starting the crawling process.

			// Create a buffered channel to limit concurrent requests.
			concurrentRequests := make(chan struct{}, 10) 

			// Create a wait group to wait for all goroutines to finish.
			var wg sync.WaitGroup

			// Create a new Colly collector with the specified settings.
			c := colly.NewCollector(
				colly.MaxDepth(0),
				colly.Async(true),
				colly.IgnoreRobotsTxt(),
			)

			// OnScraped is triggered after the program finishes scraping a resource.
			c.OnScraped(func(r *colly.Response) {
				crawlingLogs.SetText(crawlingLogs.Text + fmt.Sprintf("Finished %s\n", r.Request.URL))
				wg.Done()
			})

			// OnError is triggered if an error occurs while processing a request.
			c.OnError(func(r *colly.Response, err error) {
				crawlingLogs.SetText(crawlingLogs.Text + fmt.Sprintf("Something went wrong, https:%s\n\n", err))
				wg.Done()
			})

			// Start the crawling process by visiting the inputURLs and checking for url validity.
			for _, inputURL := range inputURLs {
				wg.Add(1)
				concurrentRequests <- struct{}{} // Acquire a slot from the buffered channel.

				// Handle the list of errors returned by checkURLValidity()
				errors := checkURLValidity(inputURL)
				if len(errors) > 0 {
					crawlingLogs.SetText(crawlingLogs.Text + fmt.Sprintf("url: %s\n", inputURL))
					for _, err := range errors {
						crawlingLogs.SetText(crawlingLogs.Text + err.Error() + "\n\n")
					}
					wg.Done()
					<-concurrentRequests // Release the slot in case of errors.
				} else {
					// If the URL is valid, visit it using colly in a separate goroutine.
					go func(url string) {
						c.Visit(url)
						<-concurrentRequests // Release the slot when the request is complete.
					}(inputURL)
				}
			}

			// Wait for all goroutines to finish their work.
			wg.Wait()
		} else {
			crawlingLogs.SetText("Please enter at least one valid URL.")
		}
	})

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Enter website urls (separate by spaces and limit to 25):", Widget: entry},
		},
	}

	clearBtn := widget.NewButtonWithIcon("Reset", theme.CancelIcon(), func(){
		crawlingLogs.SetText("")
		entry.SetText("")
	})

	myWindow.SetContent(container.NewVBox(form, crawlingLogsLabel, scrollContainer, clearBtn, startButton))
	myWindow.ShowAndRun()
}
