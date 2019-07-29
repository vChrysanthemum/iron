package iron

import "testing"

func AssertErrIsNilForTest(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
