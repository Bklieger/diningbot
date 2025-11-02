package config

const (
	DefaultBaseURL = "https://rdeapps.stanford.edu/dininghallmenu/"
)

// LocationMap maps display names to API values
var LocationMap = map[string]string{
	"Arrillaga Family Dining Commons": "Arrillaga",
	"Branner Dining":                  "Branner",
	"EVGR Dining":                     "EVGR",
	"Florence Moore Dining":           "FlorenceMoore",
	"Gerhard Casper Dining":           "GerhardCasper",
	"Lakeside Dining":                 "Lakeside",
	"Ricker Dining":                   "Ricker",
	"Stern Dining":                    "Stern",
	"Wilbur Dining":                   "Wilbur",
}

// ValidLocations contains all valid dining hall display names
var ValidLocations = []string{
	"Arrillaga Family Dining Commons",
	"Branner Dining",
	"EVGR Dining",
	"Florence Moore Dining",
	"Gerhard Casper Dining",
	"Lakeside Dining",
	"Ricker Dining",
	"Stern Dining",
	"Wilbur Dining",
}

// ValidMealTypes contains all valid meal type values
var ValidMealTypes = []string{
	"Breakfast",
	"Lunch",
	"Dinner",
	"Brunch",
}

// IsValidLocation checks if a location is valid
func IsValidLocation(location string) bool {
	_, exists := LocationMap[location]
	return exists
}

// GetLocationValue returns the API value for a display name
func GetLocationValue(displayName string) string {
	return LocationMap[displayName]
}

// IsValidMealType checks if a meal type is valid
func IsValidMealType(mealType string) bool {
	for _, mt := range ValidMealTypes {
		if mt == mealType {
			return true
		}
	}
	return false
}
