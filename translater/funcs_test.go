package translater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	assert.Equal(t, 25, len(Translators))
}
