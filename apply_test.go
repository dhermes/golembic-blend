package golembic_test

import (
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/ex"

	golembic "github.com/dhermes/golembic-blend"
)

func TestNewApplyConfig(t *testing.T) {
	it := assert.New(t)

	// Happy Path
	ac, err := golembic.NewApplyConfig()
	it.Nil(err)
	expected := &golembic.ApplyConfig{}
	it.Equal(expected, ac)

	// Sad Path
	known := ex.New("WRENCH")
	opt := func(_ *golembic.ApplyConfig) error {
		return known
	}
	ac, err = golembic.NewApplyConfig(opt)
	it.Nil(ac)
	it.Equal(known, err)
}
