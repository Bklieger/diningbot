package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDiningHallClient(t *testing.T) {
	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewDiningHallClient() returned nil client")
	}
	if client.client == nil {
		t.Error("client.client is nil")
	}
	if client.jar == nil {
		t.Error("client.jar is nil")
	}
	if client.baseURL == "" {
		t.Error("client.baseURL is empty")
	}
}

func TestSetBaseURL(t *testing.T) {
	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	testURL := "http://test.example.com/"
	client.SetBaseURL(testURL)
	if client.baseURL != testURL {
		t.Errorf("SetBaseURL() = %v, want %v", client.baseURL, testURL)
	}
}

func TestInitializeSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %v", r.Method)
		}
		html := `<html><body>
			<input type="hidden" name="__VIEWSTATE" value="test_viewstate_123" />
			<input type="hidden" name="__EVENTVALIDATION" value="test_validation_456" />
			<input type="hidden" name="__VIEWSTATEGENERATOR" value="test_generator_789" />
		</body></html>`
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	err = client.initializeSession()
	if err != nil {
		t.Fatalf("initializeSession() error = %v", err)
	}

	if client.viewState != "test_viewstate_123" {
		t.Errorf("viewState = %v, want test_viewstate_123", client.viewState)
	}
	if client.eventValidation != "test_validation_456" {
		t.Errorf("eventValidation = %v, want test_validation_456", client.eventValidation)
	}
	if client.viewStateGenerator != "test_generator_789" {
		t.Errorf("viewStateGenerator = %v, want test_generator_789", client.viewStateGenerator)
	}
}

func TestInitializeSessionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	err = client.initializeSession()
	if err == nil {
		t.Error("initializeSession() should return error on 500 status")
	}
}

func TestInitializeSessionNetworkError(t *testing.T) {
	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL("http://localhost:0/invalid")

	err = client.initializeSession()
	if err == nil {
		t.Error("initializeSession() should return error on network failure")
	}
}

func TestGetMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			html := `<html><body>
				<input type="hidden" name="__VIEWSTATE" value="initial_viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="initial_validation" />
				<input type="hidden" name="__VIEWSTATEGENERATOR" value="initial_generator" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		} else if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				t.Errorf("ParseForm error: %v", err)
			}

			if r.Form.Get("ctl00$MainContent$lstMealType") != "Breakfast" {
				t.Errorf("Expected meal type Breakfast, got %v", r.Form.Get("ctl00$MainContent$lstMealType"))
			}

			html := `<html><body>
				<table>
					<tr><td class="MenuItem">Scrambled Eggs</td></tr>
					<tr><td class="MenuItem">Bacon</td></tr>
					<tr><td class="MenuItem">Toast</td></tr>
				</table>
				<input type="hidden" name="__VIEWSTATE" value="updated_viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="updated_validation" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		}
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	foods, err := client.GetMenu("Arrillaga Family Dining Commons", "11/4/2024", "Breakfast")
	if err != nil {
		t.Fatalf("GetMenu() error = %v", err)
	}

	expected := []string{"Scrambled Eggs", "Bacon", "Toast"}
	if len(foods) != len(expected) {
		t.Errorf("GetMenu() returned %d items, want %d", len(foods), len(expected))
	}

	for _, expectedItem := range expected {
		found := false
		for _, food := range foods {
			if food == expectedItem {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetMenu() missing item %v", expectedItem)
		}
	}

	if client.viewState != "updated_viewstate" {
		t.Errorf("viewState = %v, want updated_viewstate", client.viewState)
	}
}

func TestGetMenuWithExistingSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			html := `<html><body>
				<table>
					<tr><td class="MenuItem">Oatmeal</td></tr>
				</table>
				<input type="hidden" name="__VIEWSTATE" value="viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="validation" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		}
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")
	client.viewState = "existing_viewstate"
	client.eventValidation = "existing_validation"
	client.viewStateGenerator = "existing_generator"

	foods, err := client.GetMenu("Arrillaga Family Dining Commons", "11/4/2024", "Breakfast")
	if err != nil {
		t.Fatalf("GetMenu() error = %v", err)
	}

	if len(foods) == 0 {
		t.Error("GetMenu() should return at least one item")
	}
}

func TestGetMenuInvalidLocation(t *testing.T) {
	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	_, err = client.GetMenu("Invalid Location", "11/4/2024", "Breakfast")
	if err == nil {
		t.Error("GetMenu() should return error for invalid location")
	}
}

func TestGetMenuInvalidMealType(t *testing.T) {
	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	_, err = client.GetMenu("Arrillaga Family Dining Commons", "11/4/2024", "Invalid")
	if err == nil {
		t.Error("GetMenu() should return error for invalid meal type")
	}
}

func TestGetMenuErrorOnInitialization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	_, err = client.GetMenu("Arrillaga Family Dining Commons", "11/4/2024", "Breakfast")
	if err == nil {
		t.Error("GetMenu() should return error when session initialization fails")
	}
}

func TestGetMenuErrorOnRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			html := `<html><body>
				<input type="hidden" name="__VIEWSTATE" value="viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="validation" />
				<input type="hidden" name="__VIEWSTATEGENERATOR" value="generator" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	_, err = client.GetMenu("Arrillaga Family Dining Commons", "11/4/2024", "Breakfast")
	if err == nil {
		t.Error("GetMenu() should return error when POST request fails")
	}
}

func TestGetBreakfastMenu(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			html := `<html><body>
				<input type="hidden" name="__VIEWSTATE" value="viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="validation" />
				<input type="hidden" name="__VIEWSTATEGENERATOR" value="generator" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		} else if r.Method == "POST" {
			html := `<html><body>
				<table>
					<tr><td class="MenuItem">Eggs</td></tr>
				</table>
				<input type="hidden" name="__VIEWSTATE" value="viewstate" />
				<input type="hidden" name="__EVENTVALIDATION" value="validation" />
			</body></html>`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(html))
		}
	}))
	defer server.Close()

	client, err := NewDiningHallClient()
	if err != nil {
		t.Fatalf("NewDiningHallClient() error = %v", err)
	}

	client.SetBaseURL(server.URL + "/")

	foods, err := client.GetBreakfastMenu("Arrillaga Family Dining Commons", "11/4/2024")
	if err != nil {
		t.Fatalf("GetBreakfastMenu() error = %v", err)
	}

	if len(foods) == 0 {
		t.Error("GetBreakfastMenu() should return at least one item")
	}
}
