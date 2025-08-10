package storage

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkInMemoryRepositoryGetUserURLs(b *testing.B) {
	userCounts := []int{10, 100}
	urlsPerUser := []int{10, 100}

	for _, userCount := range userCounts {
		for _, urlCount := range urlsPerUser {
			b.Run(fmt.Sprintf("users_%d_urls_%d", userCount, urlCount), func(b *testing.B) {
				repo := NewInMemoryRepository()
				ctx := context.Background()

				// Pre-populate repository
				for userID := range userCount {
					for urlID := range urlCount {
						shortID := fmt.Sprintf("user_%d_url_%d", userID, urlID)
						originalURL := fmt.Sprintf("https://user%d.example.com/%d", userID, urlID)
						repo.Add(ctx, shortID, originalURL, int64(userID))
					}
				}

				b.ResetTimer()
				for b.Loop() {
					userID := int64(rand.Intn(userCount))
					_, _ = repo.GetUserURLs(ctx, userID)
				}
			})
		}
	}
}
