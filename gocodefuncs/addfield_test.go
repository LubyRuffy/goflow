package gocodefuncs

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"regexp"
	"testing"
)

func TestAddField(t *testing.T) {

	matches := regexp.MustCompile("bbb:\\n  (.*?)\\n\\n").FindAllStringSubmatch("bbb:\n  2\n  \n\nc: 3\n", -1)
	assert.True(t, len(matches) == 0)

	matches = regexp.MustCompile("(?is)bbb:\\n  (.*?)\\n\\n").FindAllStringSubmatch("bbb:\n  2\n  \n\nc: 3\n", -1)
	assert.True(t, len(matches) > 0)

	// 一定要注意(?is)
	addRegex, err := regexp.Compile("(?is)bbb:\\n  (.*?)\\n\\n")
	assert.Nil(t, err)
	line := "{\"cert\":\"a:\\n  1\\n\\nbbb:\\n  dcm.pogogt.de\\n  map.pogogt.de\\n\\nc: 3\\n\",\"domain\":\"pogogt.de\",\"host\":\"https://proxy.pogogt.de\",\"ip\":\"91.132.145.136\",\"port\":\"443\"}"
	cert := "a:\n  1\n\nbbb:\n  dcm.pogogt.de\n  map.pogogt.de\n\nc: 3\n"
	assert.Equal(t, cert, gjson.Get(line, "cert").String())
	matches = addRegex.FindAllStringSubmatch(cert, -1)
	assert.True(t, len(matches) > 0)
	assert.True(t, len(matches[0]) > 0)

}
