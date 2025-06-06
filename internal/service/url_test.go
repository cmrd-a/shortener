package service

import (
	"context"
	"testing"

	"github.com/cmrd-a/shortener/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetOriginal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)

	value := "ya.ru"
	short := "RaNdOm"
	m.EXPECT().Get(short).Return(value, nil)

	svc := NewURLService("localhost", m)
	original, err := svc.GetOriginal(short)

	require.NoError(t, err)
	require.Equal(t, original, value)
}


func TestShortenBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)

	expected := "ya.ru"
   originals := make(map[string]string)
   originals["a"]="1"
   originals["b"]="2"
   ctx := context.Background()

	m.EXPECT().AddBatch(ctx, originals).Return(nil)

	svc := NewURLService("localhost", m)
	shorts, err := svc.ShortenBatch(ctx, originals)

	require.NoError(t, err)
	require.Equal(t, shorts, expected)
}