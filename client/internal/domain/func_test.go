package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// --- extractPrefixDSN ---
//

func TestExtractPrefixDSN(t *testing.T) {

	str := "Foo:Bar"
	want := "Foo"

	rxData := extractPrefixDSN(str)
	assert.Equalf(t, want, rxData, "Нет соответствия")
}
