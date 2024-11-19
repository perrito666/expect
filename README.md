# expect

A go test tool to manage expectations.

### Notes on the origin of the tool

Expect is born as a spiritual continuation of [abide](https://github.com/beme/abide), a tool that I have used and liked,
but is no longer maintained. I originally thought about forking it but the changes I wanted to make were so many that a
clean slate was a better idea.

### What it does

Expect is intended to allow comparing large volumes of output in tests against a pre-established snapshot. This is not
intended for expectations of functions producing actual structs but large corpuses of either text, markup languages or
other protocol dumps that are not otherwise easy to compare.

A simple mechanism is provided to maintain the snapshots up to date.

### Out of the box

We provide a set of supported types out of the box and the possibility to expand them:

* String: with simple diff output on differences
* String: with rich colored diff output on differences
* JSON: with comparison of equivalence rather than equality
* HTTP Response: with the ability to set specific comparators per ContentType
* HTTP Request (WIP): with the ability to set specific comparators per ContentType

Some aspects can be configuring by having a json file in your tests directory.

A universal mechanism to supply values as replacement of others is provided and left to be handled by the comparable
type

### Usage

There are two aspects about using this tool:

#### The flags

When running the tests (typically `go test .`) you can append, at the end, `--` and then the supported flags:

* `-u`: will update existing expectation snapshots with the passed comparable, it means one wants to set the current
  results as canon
* `-cleanup`: needs to be used along with `-u` and will delete all snapshots that are no longer use.

For `-cleanup` to work you need to also add a call to `Cleanup()` in your `TestMain` after `m.Run`

```go
package foo

import (
  "testing"

  "perri.to/expect"
)

func TestMain(m *testing.M) {
  m.Run()
  expect.Cleanup()
}
```

if the `-cleanup` flag was passed then that invocation will remove all unreferenced snapshots, otherwise it will do nothing.

You can alternatively use `expect.MustCleanup()` which will return an error (which you will need to handle) if a cleanup
was in order but not requested.

#### The configuration

##### In general
It is expected to be in json format (as seen in the example below) in a file called `expectations.json`.
Configuration will be looked for in the directory where the test file lives (we actually look up in the stack from the
call to ReadConfig within our code until we find a `_test.go` file and then look for a `expectations.json` file in the same)
A sample configuration:

```json
{
  "grouping": "by_test_file",
  "snapshot_dir": "/optional/abs/or/relative/to/store/expectations",
  "replacers": {
    "json": {
      "valueA": "replacement",
      "valueB": "another_replacement"
    },
    "string": {
      "stringA": "stringReplacement"
    }
  }
}
```

##### Per assertion

Additionally, you can use `FromSnapshotWithConfig` to pass a configuration for a single assertion, this will override the
global configuration. It is mostly useful for very particularly formatted snapshots. This function takes the configuration
as a Go struct, you can maintain it in a json file and load it with the available tools for the `Config` type

```go

Notice that this is all optional.



##### JSON

For the `json` parsing we use [`github.com/tidwall/sjson`](https://github.com/tidwall/sjson), a few notes that might
not be so clear from their documentation:
* Paths can be specified to json elements:
  * In `{"name":{"first":"Janet","last":"Prichard"},"age":47}` you can reference `name.janet`
  * For arrays, you can use the number `{"name":{"first":"Janet","last":"Prichard"},"age":47,"children":["Sara","Alex","Jack"]}` and reference `children.1`
  * You can indicate "all children" using `#`, for the above example `children.#`.
  * If the root of the `json` file is an array, you can use `#` leading `#.a_key` for example will affect the `a_key` key of all children in the array.

#### The code

There are two helpers provided to compare expectations.

```go
package foo

import (
  "testing"

  "perri.to/expect"
  "perri.to/expect/snapshots/comparabletypes"
)

func TestSomething(t *testing.T) {
  var result string
  // ... Do something
  c := comparabletypes.NewStringComparable(result)
  expect.FromSnapshot(t, "a name for our snapshot", c)
}
```

This will yield error if:

* The string in `result` does not match what we have in store
* The name of the snapshot is repeated within the package (this saves a lot of time chasing random no-match errors that
  happen due to order of tests)

Now an example with a response.

```go
package foo

import (
  "net/http"
  "testing"

  "perri.to/expect"
  "perri.to/expect/snapshots/comparabletypes"
)

func TestSomethingNetworked(t *testing.T) {
  resp, err := http.Get("https://perri.to/random/json/endpoint") // no, it does not work
  if err != nil {
    t.Fatal(err)
  }
  c, err := comparabletypes.NewResponse(resp, true)
  if err != nil {
    t.Fatal(err)
  }
  // This is built-in but is a nice sample, you can override the built ins.
  c.RegisterHandler("application/json", comparabletypes.NewJSONFromString)
  expect.FromSnapshot(t, "a request to perrito", &c)
}
```

A resulting output for a failure of a json response would be (notice also differences in status and headers)

![A sample http response difference](media/http_response_diff.jpg)

#### Examples

There are a few examples in the examples folder, these will fail and are mostly to display how it looks
