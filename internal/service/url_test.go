package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/cmrd-a/shortener/internal/service/service_mocks"
	"github.com/cmrd-a/shortener/internal/storage"
	"github.com/cmrd-a/shortener/internal/storage/storage_mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetOriginal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mr := storage_mocks.NewMockRepository(ctrl)

	value := "ya.ru"
	short := "RaNdOm"
	ctx := context.TODO()
	mr.EXPECT().Get(ctx, short).Return(value, nil)
	generator := NewShortGenerator()
	svc := NewURLService(generator, "localhost", mr)
	original, err := svc.GetOriginal(ctx, short)

	require.NoError(t, err)
	require.Equal(t, original, value)
}

func TestShortenBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mr := storage_mocks.NewMockRepository(ctrl)
	mg := service_mocks.NewMockGenerator(ctrl)
	s1 := "short1"
	s2 := "short2"
	o1 := "https://regex101.com"
	o2 := "https://www.jaegertracing.io"
	mg.EXPECT().Generate().Return(s1)
	mg.EXPECT().Generate().Return(s2)

	corOriginals := make(map[string]string, 2)
	corOriginals["cor1"] = o1
	corOriginals["cor2"] = o2
	ctx := context.TODO()
	var userID int64 = 1
	storedURLs := make([]storage.StoredURL, 0)
	storedURLs = append(storedURLs, storage.StoredURL{ShortID: s1, OriginalURL: o1, UserID: userID})
	storedURLs = append(storedURLs, storage.StoredURL{ShortID: s2, OriginalURL: o2, UserID: userID})

	mr.EXPECT().AddBatch(ctx, userID, storedURLs).Return(nil)
	svc := NewURLService(mg, "localhost", mr)
	shorts, err := svc.ShortenBatch(ctx, userID, corOriginals)

	require.NoError(t, err)
	expected := make(map[string]string)
	expected["cor1"] = fmt.Sprintf("localhost/%s", s1)
	expected["cor2"] = fmt.Sprintf("localhost/%s", s2)
	require.Equal(t, shorts, expected)
}
