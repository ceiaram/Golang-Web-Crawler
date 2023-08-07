// Refernce(s): https://www.hoodoo.digital/blog/how-to-make-a-web-crawler-using-go-and-colly-tutorial
// https://code.visualstudio.com/docs/cpp/config-mingw
package main

import (
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

func main() {
	// Create a new Fyne application instance.
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Web Crawler")

	// Entry widget to take user input.
	entry := widget.NewEntry()

	// MultiLineEntry widget to display the crawled data.
	textArea := widget.NewMultiLineEntry()

	// Button widget to start the crawling process.
	startButton := widget.NewButton("Start Crawling", func() {
		inputURLs := strings.Fields(entry.Text)
		if len(inputURLs) > 0 {
			// Clear the text area before starting the crawling process.
			textArea.SetText("")

			// Create a new Colly collector (crawler) with the specified settings.
			c := colly.NewCollector(
				colly.MaxDepth(0),           // Set the maximum depth of the crawling process (0 means no limit).
				colly.Async(true),           // Enable asynchronous scraping.
				colly.IgnoreRobotsTxt(),     // Ignore the robots.txt file.
			)

			// OnScraped is triggered after the program finishes scraping a resource.
			// It appends the URL of the finished resource to the text area.
			c.OnScraped(func(r *colly.Response) {
				textArea.SetText(textArea.Text + fmt.Sprintf("Finished %s\n", r.Request.URL))
			})

			// OnError is triggered if an error occurs while processing a request.
			c.OnError(func(r *colly.Response, err error) {
				textArea.SetText(textArea.Text + fmt.Sprintf("Something went wrong, https:%s\n", err))
			})

			// Start the crawling process by visiting the inputURLs and checking for url validity.
			for _, inputURL := range inputURLs {
				// Handle the list of errors returned by checkURLValidity()
				errors := checkURLValidity(inputURL)
				if len(errors) > 0 {
					textArea.SetText(textArea.Text + fmt.Sprintf("url: %s\n", inputURL))
					for _, err := range errors {
						textArea.SetText(textArea.Text + err.Error() + "\n")
					}
				} else {
					// If the URL is valid, visit it using colly
					c.Visit(inputURL)
				}
			}

			// Wait for the asynchronous scraping to finish.
			c.Wait()
		} else {
			// If no valid URLs are provided, display an error message.
			textArea.SetText("Please enter at least one valid URL.")
		}
	})

	// Create a form to organize the widgets.
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Enter website urls (separated by spaces):", Widget: entry},
			{Text: "Crawling Logs:", Widget: textArea},
		},
		// OnSubmit: func() {}, // Since we don't want the form to be submitted, leave this empty.
	}

	// Create a container to hold the form and the start button.
	content := container.NewVBox(form, startButton)

	// Set the content of the window to the container.
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}