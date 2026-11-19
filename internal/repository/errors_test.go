package repository

import (
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/Roxilius/gin-template/internal/domain"
)

func TestTranslate(t *testing.T) {
	other := errors.New("boom")
	tests := []struct {
		in, want error
	}{
		{nil, nil},
		{gorm.ErrRecordNotFound, domain.ErrNotFound},
		{fmt.Errorf("wrapped: %w", gorm.ErrRecordNotFound), domain.ErrNotFound},
		{gorm.ErrDuplicatedKey, domain.ErrConflict},
		{other, other},
	}
	for _, tc := range tests {
		if got := Translate(tc.in); !errors.Is(got, tc.want) {
			t.Errorf("Translate(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
