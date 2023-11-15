package comparabletypes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"perri.to/expect/snapshots"
)

var _ snapshots.Comparable = (*Response)(nil)

// Response holds comparable information of a http response.
type Response struct {
	pretty     bool
	handlers   map[string]func(string) snapshots.Comparable
	body       []byte
	headers    map[string][]string
	replacers  map[string]string
	headerKeys []string
	status     int
}

type dumpResponse struct {
	Headers map[string][]string `json:"Headers"`
	Status  int                 `json:"Status"`
}

// NewResponse returns a new instance of Response
func NewResponse(r *http.Response, pretty bool) (*Response, error) {
	rq := Response{
		pretty: pretty,
		status: r.StatusCode,
		handlers: map[string]func(string) snapshots.Comparable{
			"text/plain":       NewPrettyStringComparable,
			"application/json": NewJSONFromString,
		},
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	rq.body = b
	rq.headers = make(map[string][]string, len(r.Header))
	rq.headerKeys = make([]string, 0, len(r.Header))
	for k, v := range r.Header {
		k = strings.ToLower(k)
		_, ok := rq.headers[k]
		rq.headers[k] = append(rq.headers[k], v...)
		if !ok {
			rq.headerKeys = append(rq.headerKeys, k)
		}
	}
	// all nice and tidy for comparisons
	sort.Strings(rq.headerKeys)
	for _, k := range rq.headerKeys {
		sort.Strings(rq.headers[k])
	}
	return &rq, nil
}

func (r *Response) contentType() string {
	if ct, ok := r.headers["content-type"]; ok {
		if strings.Index(ct[0], ";") != -1 { // handle "application/json; charset=utf-8"
			return strings.Split(ct[0], ";")[0]
		}
		return ct[0]
	}
	return ""
}

func (r *Response) CompareTo(c snapshots.Comparable) (string, error) {
	cr, ok := c.(*Response)
	if !ok {
		return r.compareToString(c)
	}
	return r.compareToOtherResponse(cr)
}

func (r *Response) compareToString(c snapshots.Comparable) (string, error) {
	if r.pretty {
		return NewPrettyStringComparable(r.String()).CompareTo(NewPrettyStringComparable(c.String()))
	}
	return NewStringComparable(r.String()).CompareTo(NewStringComparable(c.String()))
}

func (r *Response) compareToOtherResponse(cr *Response) (string, error) {
	var result strings.Builder
	// Compare Status
	if r.status != cr.status {
		result.WriteString(fmt.Sprintf("Status: expected %d but got %d\n", r.status, cr.status))
	}
	// Compare Headers
	if len(r.headerKeys) != len(cr.headerKeys) {
		result.WriteString(fmt.Sprintf("Headers: expected %d Headers but got %d\n", len(r.headerKeys), len(cr.headerKeys)))
	}
	for _, k := range r.headerKeys {
		if v, ok := cr.headers[k]; !ok {
			result.WriteString(fmt.Sprintf("Headers: key %s is expected but not present\n", k))
		} else {
			if _, ok := r.replacers[k]; ok {
				// this value is replaceable, it will match
				continue
			}
			ev := r.headers[k]
			if strings.Join(ev, ", ") != strings.Join(v, ", ") {
				result.WriteString(fmt.Sprintf("Headers: key %s has value %v but we expected %v\n", k, v, ev))
			}
		}
	}
	for _, k := range cr.headerKeys {
		if _, ok := r.headers[k]; !ok {
			v := cr.headers[k]
			result.WriteString(fmt.Sprintf("Headers: key %s is not expected but present, with value %s\n", k, v))
		}
	}

	ect := r.contentType()
	ct := cr.contentType()
	handler, hasHandler := r.handlers[ect]
	if ect != ct || ect == "" || !hasHandler {
		if !reflect.DeepEqual(r.body, cr.body) {
			result.WriteString("BODY: bodies are different, please inspect them\n")
		}
		return result.String(), nil
	}

	rb := handler(string(r.body))
	crb := handler(string(cr.body))
	bdiff, err := rb.CompareTo(crb)
	if err != nil {
		return "", fmt.Errorf("comparing bodies")
	}
	if bdiff != "" {
		result.WriteString(bdiff)
	}
	return result.String(), nil
}

func (r *Response) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("STATUS: %d\n", r.status))
	for _, k := range r.headerKeys {
		var v string
		if nv, ok := r.replacers[k]; ok {
			v = nv
		} else {
			v = strings.Join(r.headers[k], ", ")
		}
		s.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	s.WriteString("\n")
	s.Write(r.body)
	return s.String()
}

func (r *Response) Kind() snapshots.Kind {
	return "http-response"
}

func (r *Response) Dump() []byte {
	dumpable := dumpResponse{
		Headers: r.headers,
		Status:  r.status,
	}
	m, err := json.MarshalIndent(&dumpable, "", "  ")
	if err != nil {
		panic(err)
	}

	// do an attempt at making this easier to read, in case the json in body is compressed and only if
	// we were asked to make it pretty
	if r.contentType() == "application/json" && r.pretty {
		tempBody := map[string]interface{}{}
		if err = json.Unmarshal(r.body, tempBody); err != nil {
			return append(m, append([]byte(headerSep), r.body...)...)
		}
		if marshaledBody, err := json.MarshalIndent(tempBody, "", "  "); err == nil {
			return append(m, append([]byte(headerSep), marshaledBody...)...)
		}
	}
	return append(m, append([]byte(headerSep), r.body...)...)
}

const headerSep = "\n\n"

func (r *Response) Load(req []byte) snapshots.Comparable {
	if len(req) == 0 {
		return &Response{}
	}
	splitLine := strings.Index(string(req), headerSep)
	if splitLine == -1 {
		panic(fmt.Errorf("cannot read a request in this file"))
	}
	dumped := &dumpResponse{}
	err := json.Unmarshal(req[:splitLine], dumped)
	if err != nil {
		panic(fmt.Errorf("unmarshaling Headers: %w", err))
	}
	headerKeys := make([]string, 0, len(dumped.Headers))
	for k := range dumped.Headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)
	for _, k := range headerKeys {
		sort.Strings(dumped.Headers[k])
	}
	newR := Response{
		headerKeys: headerKeys,
		status:     dumped.Status,
		headers:    dumped.Headers,
		body:       req[splitLine+len(headerSep):],
	}
	// we want handler parity, plus user originally will modify r
	newR.handlers = r.handlers
	return &newR
}

func (r *Response) Replace(m map[string]string) {
	r.replacers = m
}

func (r *Response) Extension() string {
	return "resp_http"
}

// RegisterHandler will store in the response comparable a handler for a content-type
func (r *Response) RegisterHandler(contentType string, h func(string) snapshots.Comparable) {
	r.handlers[contentType] = h
}
