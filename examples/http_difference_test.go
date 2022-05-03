package examples

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"perri.to/expect"
	"perri.to/expect/snapshots/comparabletypes"
)

func TestResponseFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd["foo"] = []string{"bar", "baz"}
		hd["content-type"] = []string{"application/json"}
		w.WriteHeader(http.StatusOK)
		m := map[string][]string{
			"hello": []string{"world", "universe"},
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(m); err != nil {
			t.Fatal(err)
		}
	}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := comparabletypes.NewResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}
	expect.FromSnapshot(t, "ok json http response is different.", rc)
}
