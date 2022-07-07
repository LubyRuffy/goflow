package gocodefuncs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoinFofa(t *testing.T) {
	assert.Equal(t, "abc.com", expandField("${{domain}}", `{"domain":"abc.com"}`))
	assert.Equal(t, "", expandField("${{domain}}", `{"domain1":"abc.com"}`))
	assert.Equal(t, "abc.com", expandField("${{a.domain}}", `{"a":{"domain":"abc.com"}}`))
}
