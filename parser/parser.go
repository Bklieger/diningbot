package parser

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// ExtractHiddenField extracts the value of a hidden form field from HTML
func ExtractHiddenField(htmlContent, fieldName string) string {
	// Look for name="fieldName" (might have other attributes after it)
	namePattern := `name="` + fieldName + `"`
	nameStart := strings.Index(htmlContent, namePattern)
	if nameStart == -1 {
		return ""
	}

	// Find the value=" after the name, within the same tag
	valuePattern := `value="`
	// Look for value within the next 200 characters (to stay within the same tag)
	searchEnd := nameStart + 200
	if searchEnd > len(htmlContent) {
		searchEnd = len(htmlContent)
	}
	valueStart := strings.Index(htmlContent[nameStart:searchEnd], valuePattern)
	if valueStart == -1 {
		return ""
	}

	// Adjust valueStart to absolute position and skip past 'value="'
	valueStart = nameStart + valueStart + len(valuePattern)

	// Find the closing quote
	valueEnd := strings.Index(htmlContent[valueStart:], `"`)
	if valueEnd == -1 {
		return ""
	}

	return htmlContent[valueStart : valueStart+valueEnd]
}

// ExtractViewState extracts __VIEWSTATE from HTML
func ExtractViewState(htmlContent string) string {
	return ExtractHiddenField(htmlContent, "__VIEWSTATE")
}

// ExtractEventValidation extracts __EVENTVALIDATION from HTML
func ExtractEventValidation(htmlContent string) string {
	return ExtractHiddenField(htmlContent, "__EVENTVALIDATION")
}

// ExtractViewStateGenerator extracts __VIEWSTATEGENERATOR from HTML
func ExtractViewStateGenerator(htmlContent string) string {
	return ExtractHiddenField(htmlContent, "__VIEWSTATEGENERATOR")
}

// ParseFoodItems parses HTML and extracts food menu items
func ParseFoodItems(htmlContent string, debug bool) []string {
	var foods []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return foods
	}

	seen := make(map[string]bool)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Check for td elements - try multiple strategies
			if n.Data == "td" {
				var hasMenuItemClass bool
				var classValue string
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						classValue = attr.Val
						lowerClass := strings.ToLower(attr.Val)
						// Look for various class patterns
						if strings.Contains(attr.Val, "MenuItem") ||
							strings.Contains(lowerClass, "menu") ||
							strings.Contains(lowerClass, "item") ||
							strings.Contains(lowerClass, "food") ||
							strings.Contains(lowerClass, "dish") ||
							strings.Contains(lowerClass, "entry") {
							hasMenuItemClass = true
						}
					}
				}

				// If no class match, still check td in table context (might be plain td)
				foodText := extractTextFromNode(n)
				foodName := strings.TrimSpace(foodText)

				// Strip ingredients/allergens info if present
				foodName = stripIngredientInfo(foodName)

				if hasMenuItemClass || (foodName != "" && len(foodName) > 3 && len(foodName) < 100) {
					// Filter out common non-food text
					lowerFoodName := strings.ToLower(foodName)
					// Skip if it's clearly not a food item
					skip := false
					if strings.HasPrefix(lowerFoodName, "menu") ||
						strings.HasPrefix(lowerFoodName, "breakfast") ||
						strings.HasPrefix(lowerFoodName, "lunch") ||
						strings.HasPrefix(lowerFoodName, "dinner") ||
						strings.HasPrefix(lowerFoodName, "brunch") ||
						strings.HasPrefix(lowerFoodName, "dining hall") ||
						strings.Contains(lowerFoodName, "select") ||
						strings.Contains(lowerFoodName, "choose") ||
						strings.Contains(lowerFoodName, "location") ||
						strings.Contains(lowerFoodName, "date") ||
						strings.Contains(lowerFoodName, "ingredient") ||
						strings.Contains(lowerFoodName, "allergen") ||
						strings.Contains(lowerFoodName, "allergy") ||
						strings.Contains(lowerFoodName, "made on shared") ||
						len(foodName) <= 2 {
						skip = true
					}

					// Only add if we have a class match AND not skipped
					if hasMenuItemClass && !skip && !seen[foodName] {
						// Additional check: must not look like ingredient/allergen text
						if !looksLikeIngredientText(foodName) {
							if debug {
								fmt.Printf("DEBUG: Found food item via class '%s': %s\n", classValue, foodName)
							}
							foods = append(foods, foodName)
							seen[foodName] = true
						} else if debug {
							fmt.Printf("DEBUG: Skipped ingredient-like text: %s\n", foodName)
						}
					}
				}
			}

			// Also check for div elements that might contain menu items
			if n.Data == "div" {
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						lowerClass := strings.ToLower(attr.Val)
						if strings.Contains(attr.Val, "MenuItem") ||
							strings.Contains(lowerClass, "menu-item") ||
							strings.Contains(lowerClass, "food-item") {
							foodText := extractTextFromNode(n)
							foodName := strings.TrimSpace(foodText)
							if foodName != "" && len(foodName) > 2 && !seen[foodName] {
								foods = append(foods, foodName)
								seen[foodName] = true
							}
							break
						}
					}
				}
			}

			// Check for span elements
			if n.Data == "span" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(strings.ToLower(attr.Val), "item") {
						foodText := extractTextFromNode(n)
						foodName := strings.TrimSpace(foodText)
						if foodName != "" && len(foodName) > 2 && !seen[foodName] {
							foods = append(foods, foodName)
							seen[foodName] = true
						}
						break
					}
				}
			}

			// Check for h3 elements with clsLabel_Name (main food items)
			if n.Data == "h3" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "clsLabel_Name") {
						foodText := extractTextFromNode(n)
						foodName := strings.TrimSpace(foodText)
						foodName = stripIngredientInfo(foodName)

						if foodName != "" && len(foodName) > 2 && !looksLikeIngredientText(foodName) && !seen[foodName] {
							if debug {
								fmt.Printf("DEBUG: Found food item in h3: %s\n", foodName)
							}
							foods = append(foods, foodName)
							seen[foodName] = true
						}
						break
					}
				}
			}

			// Check for list items (li) - menus are often in lists
			if n.Data == "li" {
				foodText := extractTextFromNode(n)
				foodName := strings.TrimSpace(foodText)
				foodName = stripIngredientInfo(foodName)
				lowerFoodName := strings.ToLower(foodName)
				// Filter out navigation and non-food items
				if foodName != "" && len(foodName) > 3 && len(foodName) < 100 &&
					!strings.HasPrefix(lowerFoodName, "menu") &&
					!strings.HasPrefix(lowerFoodName, "breakfast") &&
					!strings.HasPrefix(lowerFoodName, "lunch") &&
					!strings.HasPrefix(lowerFoodName, "dinner") &&
					!strings.HasPrefix(lowerFoodName, "brunch") &&
					!strings.Contains(lowerFoodName, "select") &&
					!strings.Contains(lowerFoodName, "location") &&
					!strings.Contains(lowerFoodName, "date") &&
					!strings.ContainsAny(foodName, "{}[]()|\\/") &&
					!seen[foodName] {
					if debug {
						fmt.Printf("DEBUG: Found food item in list: %s\n", foodName)
					}
					foods = append(foods, foodName)
					seen[foodName] = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return foods
}

func stripIngredientInfo(text string) string {
	// Strip everything after "Ingredients:" or "Allergens:"
	if idx := strings.Index(text, " Ingredients:"); idx != -1 {
		text = text[:idx]
	}
	if idx := strings.Index(text, " Allergens:"); idx != -1 {
		text = text[:idx]
	}
	return strings.TrimSpace(text)
}

func looksLikeIngredientText(text string) bool {
	lower := strings.ToLower(text)
	// After stripping, check if what's left is just ingredient/allergen keywords
	return lower == "ingredients" ||
		lower == "allergens" ||
		lower == "allergy" ||
		strings.HasPrefix(lower, "made on shared") ||
		(len(text) > 0 && strings.Contains(lower, "chef's choice") && !strings.Contains(lower, " "))
}

func getParentTableElement(n *html.Node) *html.Node {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && (p.Data == "table" || p.Data == "tr" || p.Data == "tbody") {
			return p
		}
	}
	return nil
}

func extractTextFromNode(n *html.Node) string {
	var textBuilder strings.Builder
	var extractText func(*html.Node)
	extractText = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				if textBuilder.Len() > 0 {
					textBuilder.WriteString(" ")
				}
				textBuilder.WriteString(text)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(n)
	return textBuilder.String()
}
