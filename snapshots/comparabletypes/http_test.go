package comparabletypes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPResponse(t *testing.T) {
	var responseBody map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd["foo"] = []string{"bar", "baz"}
		hd["content-type"] = []string{"application/json"}
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(responseBody); err != nil {
			t.Fatal(err)
		}
	}))
	defer srv.Close()

	responseBody = map[string][]string{
		"hello": []string{"world", "universe"},
	}
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := NewResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}
	rc.Replace(map[string]string{"content-length": "0"})

	responseBody = map[string][]string{
		"bye": []string{"foo", "universe"},
	}
	resp1, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	rc1, err := NewResponse(resp1, true)
	if err != nil {
		t.Fatal(err)
	}
	rc1.Replace(map[string]string{"content-length": "0"})

	const expectedDiff = `{
    [0;32m"bye": [[0m
        [0;32m"foo",[0m
        [0;32m"universe"[0m
    [0;32m][0m,
    [0;31m"hello": [[0m
        [0;31m"world",[0m
        [0;31m"universe"[0m
    [0;31m][0m
}`
	diff, err := rc.CompareTo(rc1)
	if err != nil {
		t.Fatal(err)
	}
	if diff != expectedDiff {
		t.Log("got diff")
		t.Log(diff)
		t.Log("expected diff")
		t.Log(diff)
		t.FailNow()
	}
}
