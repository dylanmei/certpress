package certpress

import (
	"testing"
)

func Test_fixup_pem_encoded_text_with_spaces(t *testing.T) {
	source := `-----HELLO WORLD----- sdfsdf sdfsdf -----HELLO WORLD-----`
	expect := `-----HELLO WORLD-----
sdfsdf
sdfsdf
-----HELLO WORLD-----`
	actual := pemFixupWhitespace(source)

	if actual != expect {
		t.Errorf("Expected %s but got %s", expect, actual)
	}
}

func Test_fixup_pem_encoded_text_without_spaces(t *testing.T) {
	source := `-----HELLO WORLD-----
sdfsdf
sdfsdf
-----HELLO WORLD-----`
	actual := pemFixupWhitespace(source)

	if actual != source {
		t.Errorf("Expected %s but got %s", source, actual)
	}
}
