package repository

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChatsLocalStorage_AddGet(t *testing.T) {
	s := NewChatsLocalStorage()

	chatID := int64(125)
	username := "username"

	err := s.Add(context.Background(), chatID, username)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 1, len(s.m))

	u, err := s.Get(context.Background(), chatID)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, username, u)
}
