package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bklieger/diningbot/config"
	"github.com/bklieger/diningbot/parser"
	"golang.org/x/net/publicsuffix"
)

type DiningHallClient struct {
	client             *http.Client
	jar                *cookiejar.Jar
	baseURL            string
	viewState          string
	eventValidation    string
	viewStateGenerator string
	Debug              bool // Enable debug output
}

func NewDiningHallClient() (*DiningHallClient, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	return &DiningHallClient{
		client:  client,
		jar:     jar,
		baseURL: config.DefaultBaseURL,
	}, nil
}

func (d *DiningHallClient) SetBaseURL(url string) {
	d.baseURL = url
}

func (d *DiningHallClient) initializeSession() error {
	// Load Menu.aspx to get the form's ViewState
	menuURL := strings.TrimSuffix(d.baseURL, "/") + "/Menu.aspx"
	req, err := http.NewRequest("GET", menuURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if d.Debug {
		fmt.Printf("DEBUG: Initial session cookies:\n")
		cookies := d.jar.Cookies(req.URL)
		for _, cookie := range cookies {
			fmt.Printf("DEBUG:   %s=%s\n", cookie.Name, cookie.Value)
		}
		// Save initial HTML to file
		filename := "debug_initial_page.html"
		if err := os.WriteFile(filename, body, 0644); err == nil {
			fmt.Printf("DEBUG: Saved initial page to %s\n", filename)
		}
	}

	// Extract ViewState and EventValidation from initial page
	htmlContent := string(body)
	d.viewState = parser.ExtractViewState(htmlContent)
	d.eventValidation = parser.ExtractEventValidation(htmlContent)
	d.viewStateGenerator = parser.ExtractViewStateGenerator(htmlContent)

	if d.Debug {
		fmt.Printf("DEBUG: Extracted ViewState length: %d\n", len(d.viewState))
		fmt.Printf("DEBUG: Extracted EventValidation length: %d\n", len(d.eventValidation))
		fmt.Printf("DEBUG: Extracted ViewStateGenerator: %s\n", d.viewStateGenerator)
		if len(d.viewState) == 0 {
			// Show a snippet to debug why extraction failed
			if len(htmlContent) > 500 {
				fmt.Printf("DEBUG: HTML snippet: %s\n", htmlContent[:500])
			} else {
				fmt.Printf("DEBUG: Full HTML: %s\n", htmlContent)
			}
		}
	}

	return nil
}

func (d *DiningHallClient) renewSession() error {
	renewURL := strings.TrimSuffix(d.baseURL, "/") + "/RenewSession.aspx"
	req, err := http.NewRequest("GET", renewURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if d.Debug {
		fmt.Printf("DEBUG: Session renewed (status %d)\n", resp.StatusCode)
		cookies := d.jar.Cookies(req.URL)
		for _, cookie := range cookies {
			fmt.Printf("DEBUG:   %s=%s\n", cookie.Name, cookie.Value)
		}
	}

	return nil
}

// GetMenu fetches the menu for a given location, date, and meal type
func (d *DiningHallClient) GetMenu(location, date, mealType string) ([]string, error) {
	// Validate inputs
	if !config.IsValidLocation(location) {
		return nil, fmt.Errorf("invalid location: %s", location)
	}
	if !config.IsValidMealType(mealType) {
		return nil, fmt.Errorf("invalid meal type: %s", mealType)
	}

	// Initialize session if not already done
	if d.viewState == "" {
		if err := d.initializeSession(); err != nil {
			return nil, fmt.Errorf("failed to initialize session: %w", err)
		}
		// Renew session to ensure it's active
		if err := d.renewSession(); err != nil {
			return nil, fmt.Errorf("failed to renew session: %w", err)
		}
		// Re-fetch the main page to get updated ViewState after session renewal
		if err := d.initializeSession(); err != nil {
			return nil, fmt.Errorf("failed to re-initialize session: %w", err)
		}
	}

	// POST to Menu.aspx, not the base URL
	menuURL := strings.TrimSuffix(d.baseURL, "/") + "/Menu.aspx"

	// Prepare form data
	formData := url.Values{}
	formData.Set("__EVENTTARGET", "GetMenulstDay")
	formData.Set("__EVENTARGUMENT", "")
	formData.Set("__VIEWSTATE", d.viewState)
	formData.Set("__VIEWSTATEGENERATOR", d.viewStateGenerator)
	formData.Set("__EVENTVALIDATION", d.eventValidation)
	formData.Set("ctl00$MainContent$lstLocations", config.GetLocationValue(location))
	formData.Set("ctl00$MainContent$lstDay", date)
	formData.Set("ctl00$MainContent$lstMealType", mealType)

	if d.Debug {
		fmt.Printf("DEBUG: Posting to %s\n", menuURL)
		fmt.Printf("DEBUG: Form data: %v\n", formData)
	}

	req, err := http.NewRequest("POST", menuURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Origin", "https://rdeapps.stanford.edu")
	req.Header.Set("Referer", d.baseURL)

	// Set cookies
	for _, cookie := range d.jar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if d.Debug {
		fmt.Printf("DEBUG: Response status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("DEBUG: Response headers:\n")
		for key, values := range resp.Header {
			fmt.Printf("DEBUG:   %s: %v\n", key, values)
		}
		fmt.Printf("DEBUG: Cookies received:\n")
		cookies := d.jar.Cookies(req.URL)
		for _, cookie := range cookies {
			fmt.Printf("DEBUG:   %s=%s\n", cookie.Name, cookie.Value)
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if d.Debug {
			fmt.Printf("DEBUG: Error response body: %s\n", string(body))
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Update ViewState and EventValidation from response
	d.viewState = parser.ExtractViewState(string(body))
	d.eventValidation = parser.ExtractEventValidation(string(body))

	// Parse HTML to extract food items
	htmlContent := string(body)
	if d.Debug {
		fmt.Printf("DEBUG: Response HTML length: %d bytes\n", len(htmlContent))
		fmt.Printf("DEBUG: Looking for menu items...\n")

		// Save HTML to file for inspection
		filename := fmt.Sprintf("debug_response_%s_%s.html", strings.ReplaceAll(location, " ", "_"), strings.ReplaceAll(date, "/", "_"))
		if err := os.WriteFile(filename, body, 0644); err == nil {
			fmt.Printf("DEBUG: Saved HTML response to %s\n", filename)
		}

		// Print first 5000 chars for quick inspection
		if len(htmlContent) > 5000 {
			fmt.Printf("DEBUG: HTML preview (first 5000 chars):\n%s\n", htmlContent[:5000])
		} else {
			fmt.Printf("DEBUG: Full HTML response:\n%s\n", htmlContent)
		}
	}
	foods := parser.ParseFoodItems(htmlContent, d.Debug)
	if d.Debug {
		fmt.Printf("DEBUG: Found %d food items\n", len(foods))
		if len(foods) == 0 {
			// Try to find any potential menu-related elements
			if strings.Contains(htmlContent, "MenuItem") {
				fmt.Printf("DEBUG: HTML contains 'MenuItem' class\n")
			}
			if strings.Contains(htmlContent, "menu") {
				fmt.Printf("DEBUG: HTML contains 'menu' text\n")
			}
			// Also check for common patterns
			if strings.Contains(htmlContent, "td") {
				fmt.Printf("DEBUG: HTML contains table cells\n")
			}
			if strings.Contains(htmlContent, "li") {
				fmt.Printf("DEBUG: HTML contains list items\n")
			}
			if strings.Contains(htmlContent, "<div") {
				fmt.Printf("DEBUG: HTML contains div elements\n")
			}
			// Check for error messages
			if strings.Contains(strings.ToLower(htmlContent), "error") ||
				strings.Contains(strings.ToLower(htmlContent), "not found") ||
				strings.Contains(strings.ToLower(htmlContent), "no menu") {
				fmt.Printf("DEBUG: HTML appears to contain error messages\n")
			}
		}
	}
	return foods, nil
}

// GetBreakfastMenu is a convenience method for getting breakfast menus
// Deprecated: Use GetMenu with "Breakfast" as mealType instead
func (d *DiningHallClient) GetBreakfastMenu(location, date string) ([]string, error) {
	return d.GetMenu(location, date, "Breakfast")
}
