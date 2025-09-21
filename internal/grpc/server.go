package grpc

import (
	"context"
	"fmt"

	pb "github.com/cmrd-a/shortener/api/proto/shortener"
	"github.com/cmrd-a/shortener/internal/server/middleware"
	"github.com/cmrd-a/shortener/internal/service"
	"github.com/cmrd-a/shortener/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the gRPC ShortenerService
type Server struct {
	pb.UnimplementedShortenerServiceServer
	service Servicer
}

// Servicer interface matches the one from HTTP handlers
type Servicer interface {
	// Сокращает ссылку
	Shorten(ctx context.Context, original string, userID int64) (short string, err error)
	// Сокращает ссылки
	ShortenBatch(ctx context.Context, userID int64, corrOrig map[string]string) (corrShort map[string]string, err error)
	//Возвращает оригинальную ссылку
	GetOriginal(ctx context.Context, short string) (original string, err error)
	// Проверяет соединение с базой данных
	Ping(ctx context.Context) (err error)
	// Возвращает все ссылки пользователя
	GetUserURLs(ctx context.Context, userID int64) (urls []service.SvcURL, err error)
	// Удаляет ссылки пользователя
	DeleteUserURLs(ctx context.Context, userID int64, shortIDs ...string)
	// Возвращает статистику сервиса
	GetStats(ctx context.Context) (stats storage.Stats, err error)
}

// NewServer creates a new gRPC server instance
func NewServer(svc Servicer) *Server {
	return &Server{
		service: svc,
	}
}

// AddLink implements the AddLink gRPC method (POST /)
func (s *Server) AddLink(ctx context.Context, req *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is empty")
	}

	userID := middleware.GetUserID(ctx)
	shortLink, err := s.service.Shorten(ctx, req.Url, userID)

	var alreadyExistError *service.OriginalExistError
	if service.IsOriginalExistError(err, &alreadyExistError) {
		return &pb.AddLinkResponse{
			ShortUrl:      alreadyExistError.Short,
			AlreadyExists: true,
		}, nil
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddLinkResponse{
		ShortUrl:      shortLink,
		AlreadyExists: false,
	}, nil
}

// GetLink implements the GetLink gRPC method (GET /{linkId})
func (s *Server) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	if req.LinkId == "" {
		return nil, status.Error(codes.InvalidArgument, "Link ID is empty")
	}

	original, err := s.service.GetOriginal(ctx, req.LinkId)
	if err != nil {
		if storage.IsURLDeleted(err) {
			return &pb.GetLinkResponse{
				OriginalUrl: "",
				IsDeleted:   true,
			}, nil
		}
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetLinkResponse{
		OriginalUrl: original,
		IsDeleted:   false,
	}, nil
}

// Ping implements the Ping gRPC method (GET /ping)
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.service.Ping(ctx)
	if err != nil {
		return &pb.PingResponse{Healthy: false}, nil
	}
	return &pb.PingResponse{Healthy: true}, nil
}

// Shorten implements the Shorten gRPC method (POST /api/shorten)
func (s *Server) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is empty")
	}

	userID := middleware.GetUserID(ctx)
	shortLink, err := s.service.Shorten(ctx, req.Url, userID)

	var alreadyExistError *service.OriginalExistError
	if service.IsOriginalExistError(err, &alreadyExistError) {
		return &pb.ShortenResponse{
			Result:        alreadyExistError.Short,
			AlreadyExists: true,
		}, nil
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ShortenResponse{
		Result:        shortLink,
		AlreadyExists: false,
	}, nil
}

// ShortenBatch implements the ShortenBatch gRPC method (POST /api/shorten/batch)
func (s *Server) ShortenBatch(ctx context.Context, req *pb.ShortenBatchRequest) (*pb.ShortenBatchResponse, error) {
	if len(req.Items) == 0 {
		return &pb.ShortenBatchResponse{Items: []*pb.ShortenBatchResponseItem{}}, nil
	}

	// Validate and convert request
	corrOrig := make(map[string]string, len(req.Items))
	for _, item := range req.Items {
		if item.OriginalUrl == "" || item.CorrelationId == "" {
			return nil, status.Error(codes.InvalidArgument, "original_url or correlation_id is empty")
		}
		if _, exists := corrOrig[item.CorrelationId]; exists {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("duplicated correlation_id: %s", item.CorrelationId))
		}
		// Check for duplicate URLs
		for _, original := range corrOrig {
			if original == item.OriginalUrl {
				return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("duplicate original_url: %s", item.OriginalUrl))
			}
		}
		corrOrig[item.CorrelationId] = item.OriginalUrl
	}

	userID := middleware.GetUserID(ctx)
	corrShort, err := s.service.ShortenBatch(ctx, userID, corrOrig)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert response
	items := make([]*pb.ShortenBatchResponseItem, 0, len(corrShort))
	for corrID, short := range corrShort {
		items = append(items, &pb.ShortenBatchResponseItem{
			CorrelationId: corrID,
			ShortUrl:      short,
		})
	}

	return &pb.ShortenBatchResponse{Items: items}, nil
}

// GetUserURLs implements the GetUserURLs gRPC method (GET /api/user/urls)
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "User not authenticated")
	}

	urls, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(urls) == 0 {
		return &pb.GetUserURLsResponse{Urls: []*pb.UserURL{}}, nil
	}

	// Convert response
	userURLs := make([]*pb.UserURL, len(urls))
	for i, u := range urls {
		userURLs[i] = &pb.UserURL{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		}
	}

	return &pb.GetUserURLsResponse{Urls: userURLs}, nil
}

// DeleteUserURLs implements the DeleteUserURLs gRPC method (DELETE /api/user/urls)
func (s *Server) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return nil, status.Error(codes.Unauthenticated, "User not authenticated")
	}

	if len(req.ShortIds) == 0 {
		return &pb.DeleteUserURLsResponse{Success: true}, nil
	}

	s.service.DeleteUserURLs(ctx, userID, req.ShortIds...)
	return &pb.DeleteUserURLsResponse{Success: true}, nil
}

// GetStats implements the GetStats gRPC method (GET /api/internal/stats)
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}
