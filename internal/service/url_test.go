package service

import (
	"context"
	"testing"

	"github.com/cmrd-a/shortener/internal/service/service_mocks"
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
	mr.EXPECT().Get(short).Return(value, nil)
	generator := NewShortGenerator()
	svc := NewURLService(generator, "localhost", mr)
	original, err := svc.GetOriginal(short)

	require.NoError(t, err)
	require.Equal(t, original, value)
}

func TestShortenBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mr := storage_mocks.NewMockRepository(ctrl)
	mg := service_mocks.NewMockGenerator(ctrl)
	mg.EXPECT().Generate().Return("short1")
	mg.EXPECT().Generate().Return("short2")

	corOriginals := make(map[string]string)
	corOriginals["cor1"] = "https://regex101.com/"
	corOriginals["cor2"] = "https://www.jaegertracing.io/"
	ctx := context.Background()

	shortOriginals := make(map[string]string)
	shortOriginals["short1"] = "https://regex101.com/"
	shortOriginals["short2"] = "https://www.jaegertracing.io/"
	var userID int64 = 1
	mr.EXPECT().AddBatch(ctx, userID, shortOriginals).Return(nil)
	svc := NewURLService(mg, "localhost", mr)
	shorts, err := svc.ShortenBatch(ctx, userID, corOriginals)

	require.NoError(t, err)
	expected := make(map[string]string)
	expected["cor1"] = "localhost/short1"
	expected["cor2"] = "localhost/short2"
	require.Equal(t, shorts, expected)
}
