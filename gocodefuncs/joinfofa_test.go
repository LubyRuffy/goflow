package gocodefuncs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoinFofa(t *testing.T) {
	assert.Equal(t, "abc.com", ExpendVarWithJsonLine(nil, "${{domain}}", `{"domain":"abc.com"}`))
	assert.Equal(t, "", ExpendVarWithJsonLine(nil, "${{domain}}", `{"domain1":"abc.com"}`))
	assert.Equal(t, "abc.com", ExpendVarWithJsonLine(nil, "${{a.domain}}", `{"a":{"domain":"abc.com"}}`))
}
