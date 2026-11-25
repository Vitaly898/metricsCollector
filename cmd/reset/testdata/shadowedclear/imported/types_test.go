package imported

import "testing"

func TestGeneratedImportDoesNotShadowClear(t *testing.T) {
	value := Container{Entries: map[string]int{"key": 1}}
	value.Plain.Value = 9
	alias := value.Entries
	value.Reset()
	if len(value.Entries) != 0 || len(alias) != 0 || value.Plain.Value != 0 {
		t.Fatalf("reset was affected by the package named clear: %+v", value)
	}
}
