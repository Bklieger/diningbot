package parser

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseHTML(htmlStr string) (*html.Node, error) {
	return html.Parse(strings.NewReader(htmlStr))
}

func TestExtractHiddenField(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		fieldName string
		want      string
	}{
		{
			name:      "normal viewstate",
			html:      `<input type="hidden" name="__VIEWSTATE" value="test_value_123" />`,
			fieldName: "__VIEWSTATE",
			want:      "test_value_123",
		},
		{
			name:      "event validation",
			html:      `<input name="__EVENTVALIDATION" value="validation_token" />`,
			fieldName: "__EVENTVALIDATION",
			want:      "validation_token",
		},
		{
			name:      "not found",
			html:      `<html><body>No viewstate here</body></html>`,
			fieldName: "__VIEWSTATE",
			want:      "",
		},
		{
			name:      "empty value",
			html:      `<input name="__VIEWSTATE" value="" />`,
			fieldName: "__VIEWSTATE",
			want:      "",
		},
		{
			name:      "with special characters",
			html:      `<input name="__VIEWSTATE" value="encoded%2Fvalue+test" />`,
			fieldName: "__VIEWSTATE",
			want:      "encoded%2Fvalue+test",
		},
		{
			name:      "multiple fields, get correct one",
			html:      `<input name="__VIEWSTATE" value="wrong" /><input name="__EVENTVALIDATION" value="correct" />`,
			fieldName: "__EVENTVALIDATION",
			want:      "correct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHiddenField(tt.html, tt.fieldName)
			if got != tt.want {
				t.Errorf("ExtractHiddenField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractViewState(t *testing.T) {
	html := `<input name="__VIEWSTATE" value="test_viewstate_123" />`
	got := ExtractViewState(html)
	if got != "test_viewstate_123" {
		t.Errorf("ExtractViewState() = %v, want test_viewstate_123", got)
	}
}

func TestExtractEventValidation(t *testing.T) {
	html := `<input name="__EVENTVALIDATION" value="test_validation_456" />`
	got := ExtractEventValidation(html)
	if got != "test_validation_456" {
		t.Errorf("ExtractEventValidation() = %v, want test_validation_456", got)
	}
}

func TestExtractViewStateGenerator(t *testing.T) {
	html := `<input name="__VIEWSTATEGENERATOR" value="generator_789" />`
	got := ExtractViewStateGenerator(html)
	if got != "generator_789" {
		t.Errorf("ExtractViewStateGenerator() = %v, want generator_789", got)
	}
}

func TestExtractTextFromNode(t *testing.T) {
	htmlStr := `<div>Simple text</div>`
	doc, err := parseHTML(htmlStr)
	if err != nil {
		t.Fatalf("parseHTML error: %v", err)
	}

	// Navigate to the div element
	var divNode *html.Node
	var findDiv func(*html.Node)
	findDiv = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			divNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDiv(c)
		}
	}
	findDiv(doc)

	if divNode == nil {
		t.Fatal("Could not find div node")
	}

	text := extractTextFromNode(divNode)
	if !strings.Contains(text, "Simple text") {
		t.Errorf("extractTextFromNode() = %v, want to contain Simple text", text)
	}
}

func TestParseFoodItems(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		want    []string
		wantMin int
	}{
		{
			name: "td with MenuItem class",
			html: `<table><tr><td class="MenuItem">Scrambled Eggs</td></tr></table>`,
			want: []string{"Scrambled Eggs"},
		},
		{
			name: "multiple menu items",
			html: `<table>
				<tr><td class="MenuItem">Scrambled Eggs</td></tr>
				<tr><td class="MenuItem">Bacon</td></tr>
				<tr><td class="MenuItem">Toast</td></tr>
			</table>`,
			want: []string{"Scrambled Eggs", "Bacon", "Toast"},
		},
		{
			name: "div with menu-item class",
			html: `<div class="menu-item">Pancakes</div>`,
			want: []string{"Pancakes"},
		},
		{
			name: "span with item class",
			html: `<span class="food-item">Waffles</span>`,
			want: []string{"Waffles"},
		},
		{
			name: "td with food class",
			html: `<table><tr><td class="food">Oatmeal</td></tr></table>`,
			want: []string{"Oatmeal"},
		},
		{
			name: "filter out menu/breakfast text",
			html: `<table>
				<tr><td class="MenuItem">Breakfast Menu</td></tr>
				<tr><td class="MenuItem">Eggs</td></tr>
			</table>`,
			want: []string{"Eggs"},
		},
		{
			name: "filter short items",
			html: `<table>
				<tr><td class="MenuItem">Ab</td></tr>
				<tr><td class="MenuItem">Eggs</td></tr>
			</table>`,
			want: []string{"Eggs"},
		},
		{
			name: "no menu items",
			html: `<html><body><div>No menu here</div></body></html>`,
			want: []string{},
		},
		{
			name: "duplicate items",
			html: `<table>
				<tr><td class="MenuItem">Eggs</td></tr>
				<tr><td class="MenuItem">Eggs</td></tr>
			</table>`,
			want: []string{"Eggs"},
		},
		{
			name: "nested text",
			html: `<table><tr><td class="MenuItem">Scrambled <strong>Eggs</strong> with <em>Cheese</em></td></tr></table>`,
			want: []string{"Scrambled Eggs with Cheese"},
		},
		{
			name: "mixed case classes",
			html: `<div class="MenuItem">Item1</div><div class="menu-item">Item2</div><div class="food-item">Item3</div>`,
			want: []string{"Item1", "Item2", "Item3"},
		},
		{
			name: "empty text",
			html: `<td class="MenuItem"></td>`,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFoodItems(tt.html, false)
			if tt.wantMin > 0 {
				if len(got) < tt.wantMin {
					t.Errorf("ParseFoodItems() returned %d items, want at least %d", len(got), tt.wantMin)
				}
			} else {
				if len(got) != len(tt.want) {
					t.Errorf("ParseFoodItems() returned %d items, want %d", len(got), len(tt.want))
				}
				for _, item := range tt.want {
					found := false
					for _, g := range got {
						if g == item {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ParseFoodItems() missing item %v", item)
					}
				}
			}
		})
	}
}

func TestParseFoodItemsInvalidHTML(t *testing.T) {
	invalidHTML := `<html><body><div><td class="MenuItem">Eggs</td></body></html>`
	result := ParseFoodItems(invalidHTML, false)
	// ParseFoodItems handles errors gracefully - html.Parse is very lenient
	// so even "invalid" HTML will typically parse
	// This test just ensures the function doesn't panic on unusual HTML structures
	_ = result // Just verify the function completes without panicking
}

func TestParseFoodItemsComplexHTML(t *testing.T) {
	html := `<html>
		<head><title>Menu</title></head>
		<body>
			<div class="menu-container">
				<h2>Breakfast Menu</h2>
				<table class="menu-table">
					<tr>
						<td class="MenuItem">Scrambled Eggs</td>
						<td class="MenuItem">Bacon</td>
					</tr>
					<tr>
						<td class="MenuItem">French Toast</td>
						<td class="MenuItem">Hash Browns</td>
					</tr>
				</table>
				<div class="menu-item">Fresh Fruit</div>
				<span class="food-item">Orange Juice</span>
			</div>
		</body>
	</html>`

	foods := ParseFoodItems(html, false)
	expectedItems := []string{"Scrambled Eggs", "Bacon", "French Toast", "Hash Browns", "Fresh Fruit", "Orange Juice"}

	if len(foods) < len(expectedItems) {
		t.Errorf("ParseFoodItems() returned %d items, want at least %d", len(foods), len(expectedItems))
	}

	for _, expected := range expectedItems {
		found := false
		for _, food := range foods {
			if food == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ParseFoodItems() missing expected item: %v", expected)
		}
	}
}

func TestParseFoodItemsWithWhitespace(t *testing.T) {
	html := `<table><tr><td class="MenuItem">
		Scrambled Eggs
		with
		Cheese
	</td></tr></table>`

	foods := ParseFoodItems(html, false)
	if len(foods) == 0 {
		t.Error("ParseFoodItems() should handle whitespace properly")
		return
	}

	foodText := foods[0]
	if !strings.Contains(foodText, "Scrambled") || !strings.Contains(foodText, "Eggs") {
		t.Errorf("ParseFoodItems() should preserve food name, got %v", foodText)
	}
}
