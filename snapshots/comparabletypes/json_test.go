package comparabletypes

import (
	"encoding/json"
	"fmt"
	"testing"

	"perri.to/expect/snapshots"
)

func TestJSON_CompareTo(t *testing.T) {
	type fields struct {
		rawJSON json.RawMessage
	}
	type args struct {
		c snapshots.Comparable
	}
	tests := []struct {
		name     string
		fields   fields
		replacer map[string]string
		args     args
		want     string
		wantErr  bool
	}{
		{
			name: "equal",
			fields: fields{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "popup": {
    "menuitem": [
      {"value": "New", "onclick": "CreateNewDoc()"},
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)},
			args: args{c: &JSON{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "popup": {
    "menuitem": [
      {"value": "New", "onclick": "CreateNewDoc()"},
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)}},
			want: `""`,
		},
		{
			name: "equal_replaced",
			replacer: map[string]string{
				"menu.top_replaceable":       "replaced",
				"menu.deep.deep.replaceable": "also_replaced",
			},
			fields: fields{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "top_replaceable": "value.orig",
  "deep": {"deep": {"replaceable": "value1.orig"}},
  "popup": {
    "menuitem": [
      {"value": "New", "onclick": "CreateNewDoc()"},
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)},
			args: args{c: &JSON{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "top_replaceable": "value_diff.orig",
  "deep": {"deep": {"replaceable": "value1_diff.orig"}},
  "popup": {
    "menuitem": [
      {"value": "New", "onclick": "CreateNewDoc()"},
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)}},
			want: `""`,
		},
		{
			name: "half different",
			fields: fields{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "popup": {
    "menuitem": [
      {"value": "New", "onclick": "CreateNewDoc()"},
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)},
			args: args{c: &JSON{rawJSON: []byte(`{"menu": {
  "id": "file",
  "value": "File",
  "popup": {
    "menuitem": [
      {"value": "Open", "onclick": "OpenDoc()"},
      {"value": "Close", "onclick": "CloseDoc()"}
    ]
  }
}}`)}},
			want: `"{\n    \"menu\": {\n        \"id\": \"file\",\n        \"popup\": {\n            \"menuitem\": [\n                {\n                    \"onclick\": \x1b[0;33m\"CreateNewDoc()\" => \"OpenDoc()\"\x1b[0m,\n                    \"value\": \x1b[0;33m\"New\" => \"Open\"\x1b[0m\n                },\n                {\n                    \"onclick\": \x1b[0;33m\"OpenDoc()\" => \"CloseDoc()\"\x1b[0m,\n                    \"value\": \x1b[0;33m\"Open\" => \"Close\"\x1b[0m\n                },\n                \x1b[0;31m{\x1b[0m\n                    \x1b[0;31m\"onclick\": \"CloseDoc()\",\x1b[0m\n                    \x1b[0;31m\"value\": \"Close\"\x1b[0m\n                \x1b[0;31m}\x1b[0m\n            ]\n        },\n        \"value\": \"File\"\n    }\n}"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JSON{
				rawJSON: tt.fields.rawJSON,
			}
			if tt.replacer != nil {
				j.Replace(tt.replacer)
				tt.args.c.Replace(tt.replacer)
			}
			got, err := j.CompareTo(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if fmt.Sprintf("%q", got) != tt.want {
				fmt.Println(got)
				t.Errorf("CompareTo() got = \n%q\n, want \n%v", got, tt.want)
			}

		})
	}
}
