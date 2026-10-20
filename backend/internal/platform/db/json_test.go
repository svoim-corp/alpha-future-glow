package db

import "testing"

func TestJSON(t *testing.T) {
	var j JSON[map[string]int]
	if err := j.Scan([]byte(`{"a":1}`)); err != nil || j.V["a"] != 1 {
		t.Fatalf("scan: %v %v", err, j.V)
	}
	v, err := JSON[[]string]{V: []string{"x"}}.Value()
	if err != nil || v != `["x"]` {
		t.Fatalf("value: %v %v", v, err)
	}
	if err := j.Scan(nil); err != nil || j.V != nil {
		t.Fatal("nil scan")
	}
}
