package grpc

import (
	"context"
	"testing"

	pb "github.com/cmrd-a/shortener/api/proto/shortener"
	"github.com/cmrd-a/shortener/internal/server/middleware"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_AddLink(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	tests := []struct {
		name     string
		request  *pb.AddLinkRequest
		wantCode codes.Code
		wantErr  bool
	}{
		{
			name:     "successful link creation",
			request:  &pb.AddLinkRequest{Url: "https://example.com"},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:     "empty URL",
			request:  &pb.AddLinkRequest{Url: ""},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), middleware.UserIDKey, int64(123))

			resp, err := server.AddLink(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.ShortUrl)
				assert.Contains(t, resp.ShortUrl, "http://localhost:8080/")
				assert.False(t, resp.AlreadyExists)
			}
		})
	}
}

func TestServer_GetLink(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	// Add a link first
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, int64(123))
	addResp, err := server.AddLink(ctx, &pb.AddLinkRequest{Url: "https://example.com"})
	require.NoError(t, err)

	// Extract short ID from the full URL
	shortURL := addResp.ShortUrl
	shortID := shortURL[len("http://localhost:8080/"):]

	tests := []struct {
		name     string
		request  *pb.GetLinkRequest
		wantCode codes.Code
		wantErr  bool
		wantURL  string
	}{
		{
			name:     "successful link retrieval",
			request:  &pb.GetLinkRequest{LinkId: shortID},
			wantCode: codes.OK,
			wantErr:  false,
			wantURL:  "https://example.com",
		},
		{
			name:     "empty link ID",
			request:  &pb.GetLinkRequest{LinkId: ""},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
		{
			name:     "non-existent link ID",
			request:  &pb.GetLinkRequest{LinkId: "nonexistent"},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetLink(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.wantURL, resp.OriginalUrl)
				assert.False(t, resp.IsDeleted)
			}
		})
	}
}

func TestServer_Ping(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	ctx := context.Background()
	resp, err := server.Ping(ctx, &pb.PingRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Healthy)
}

func TestServer_Shorten(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	tests := []struct {
		name     string
		request  *pb.ShortenRequest
		wantCode codes.Code
		wantErr  bool
	}{
		{
			name:     "successful shorten",
			request:  &pb.ShortenRequest{Url: "https://example.com"},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:     "empty URL",
			request:  &pb.ShortenRequest{Url: ""},
			wantCode: codes.InvalidArgument,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), middleware.UserIDKey, int64(123))

			resp, err := server.Shorten(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotEmpty(t, resp.Result)
				assert.Contains(t, resp.Result, "http://localhost:8080/")
				assert.False(t, resp.AlreadyExists)
			}
		})
	}
}

func TestServer_GetUserURLs(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	// Add some URLs for user
	userID := int64(123)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	_, err := server.AddLink(ctx, &pb.AddLinkRequest{Url: "https://example1.com"})
	require.NoError(t, err)
	_, err = server.AddLink(ctx, &pb.AddLinkRequest{Url: "https://example2.com"})
	require.NoError(t, err)

	tests := []struct {
		name     string
		ctx      context.Context
		wantCode codes.Code
		wantErr  bool
		wantURLs int
	}{
		{
			name:     "successful get user URLs",
			ctx:      ctx,
			wantCode: codes.OK,
			wantErr:  false,
			wantURLs: 2,
		},
		{
			name:     "no authentication",
			ctx:      context.Background(),
			wantCode: codes.Unauthenticated,
			wantErr:  true,
			wantURLs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.GetUserURLs(tt.ctx, &pb.GetUserURLsRequest{})

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Len(t, resp.Urls, tt.wantURLs)

				for _, url := range resp.Urls {
					assert.NotEmpty(t, url.ShortUrl)
					assert.NotEmpty(t, url.OriginalUrl)
					assert.Contains(t, url.ShortUrl, "http://localhost:8080/")
				}
			}
		})
	}
}

func TestServer_DeleteUserURLs(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	// Add a URL for user
	userID := int64(123)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, userID)
	addResp, err := server.AddLink(ctx, &pb.AddLinkRequest{Url: "https://example.com"})
	require.NoError(t, err)

	// Extract short ID from the full URL
	shortURL := addResp.ShortUrl
	shortID := shortURL[len("http://localhost:8080/"):]

	tests := []struct {
		name     string
		ctx      context.Context
		request  *pb.DeleteUserURLsRequest
		wantCode codes.Code
		wantErr  bool
	}{
		{
			name: "successful delete",
			ctx:  ctx,
			request: &pb.DeleteUserURLsRequest{
				ShortIds: []string{shortID},
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "no authentication",
			ctx:  context.Background(),
			request: &pb.DeleteUserURLsRequest{
				ShortIds: []string{shortID},
			},
			wantCode: codes.Unauthenticated,
			wantErr:  true,
		},
		{
			name: "empty list",
			ctx:  ctx,
			request: &pb.DeleteUserURLsRequest{
				ShortIds: []string{},
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.DeleteUserURLs(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestServer_GetStats(t *testing.T) {
	// Create isolated repository and service
	repo := storage.NewInMemoryRepository()
	generator := service.NewShortGenerator()
	svc := service.NewURLService(generator, "http://localhost:8080", repo)
	server := NewServer(svc)

	// Add some URLs from different users
	ctx1 := context.WithValue(context.Background(), middleware.UserIDKey, int64(123))
	ctx2 := context.WithValue(context.Background(), middleware.UserIDKey, int64(456))

	_, err := server.AddLink(ctx1, &pb.AddLinkRequest{Url: "https://example1.com"})
	require.NoError(t, err)
	_, err = server.AddLink(ctx1, &pb.AddLinkRequest{Url: "https://example2.com"})
	require.NoError(t, err)
	_, err = server.AddLink(ctx2, &pb.AddLinkRequest{Url: "https://example3.com"})
	require.NoError(t, err)

	resp, err := server.GetStats(context.Background(), &pb.GetStatsRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int32(3), resp.Urls)
	assert.Equal(t, int32(2), resp.Users)
}
