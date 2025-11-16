package cache

import (
	"testing"
	"time"
)

func TestMenuCache_GetSet(t *testing.T) {
	cache := NewMenuCache(1 * time.Hour)

	// Test cache miss
	items, found := cache.Get("Location1", "1/1/2025", "Lunch")
	if found {
		t.Error("Expected cache miss, got cache hit")
	}
	if items != nil {
		t.Error("Expected nil items on cache miss")
	}

	// Test cache set and get
	expectedItems := []string{"Item1", "Item2", "Item3"}
	cache.Set("Location1", "1/1/2025", "Lunch", expectedItems)

	items, found = cache.Get("Location1", "1/1/2025", "Lunch")
	if !found {
		t.Error("Expected cache hit, got cache miss")
	}
	if len(items) != len(expectedItems) {
		t.Errorf("Expected %d items, got %d", len(expectedItems), len(items))
	}
	for i, item := range expectedItems {
		if items[i] != item {
			t.Errorf("Item %d: expected %s, got %s", i, item, items[i])
		}
	}
}

func TestMenuCache_Expiry(t *testing.T) {
	cache := NewMenuCache(100 * time.Millisecond) // Very short TTL for testing

	items := []string{"Item1", "Item2"}
	cache.Set("Location1", "1/1/2025", "Lunch", items)

	// Should be in cache immediately
	_, found := cache.Get("Location1", "1/1/2025", "Lunch")
	if !found {
		t.Error("Expected cache hit immediately after set")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("Location1", "1/1/2025", "Lunch")
	if found {
		t.Error("Expected cache miss after expiry, got cache hit")
	}
}

func TestMenuCache_Clear(t *testing.T) {
	cache := NewMenuCache(1 * time.Hour)

	cache.Set("Location1", "1/1/2025", "Lunch", []string{"Item1"})
	cache.Set("Location2", "1/1/2025", "Dinner", []string{"Item2"})

	// Verify items are cached
	_, found1 := cache.Get("Location1", "1/1/2025", "Lunch")
	_, found2 := cache.Get("Location2", "1/1/2025", "Dinner")
	if !found1 || !found2 {
		t.Error("Expected both items to be cached")
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	_, found1 = cache.Get("Location1", "1/1/2025", "Lunch")
	_, found2 = cache.Get("Location2", "1/1/2025", "Dinner")
	if found1 || found2 {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestMenuCache_CleanExpired(t *testing.T) {
	cache := NewMenuCache(100 * time.Millisecond)

	// Add items with different timestamps
	cache.Set("Location1", "1/1/2025", "Lunch", []string{"Item1"})
	time.Sleep(50 * time.Millisecond)
	cache.Set("Location2", "1/1/2025", "Dinner", []string{"Item2"})

	// Wait for first item to expire
	time.Sleep(60 * time.Millisecond)

	// Clean expired
	cache.CleanExpired()

	// First item should be gone, second should still be there
	_, found1 := cache.Get("Location1", "1/1/2025", "Lunch")
	_, found2 := cache.Get("Location2", "1/1/2025", "Dinner")

	if found1 {
		t.Error("Expected expired item to be removed")
	}
	if !found2 {
		t.Error("Expected non-expired item to remain")
	}
}

func TestMenuCache_ThreadSafety(t *testing.T) {
	cache := NewMenuCache(1 * time.Hour)

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			location := "Location1"
			date := "1/1/2025"
			mealType := "Lunch"
			items := []string{"Item1", "Item2"}

			cache.Set(location, date, mealType, items)
			_, found := cache.Get(location, date, mealType)
			if !found {
				t.Errorf("Goroutine %d: expected cache hit", id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
