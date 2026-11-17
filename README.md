# go-default

Yet another Golang's struct filling library. `go-default` fills struct fields with default values declared in struct tags. Instead of writing boilerplate constructor code to initialize configuration structs, request payloads, or any other struct with sensible defaults, you annotate the fields with `default:"..."` tags and let the library populate them — either at runtime via reflection, or ahead of time via code generation that avoids reflection entirely.

## Features

- Struct-tag driven defaults (`default:"..."`) for strings, all sized int/uint types, floats, booleans, slices/arrays, `[]byte` (base64-encoded), and `time.Duration`.
- Only zero-valued fields are filled — values already set by the caller are preserved.
- Unexported fields are skipped automatically.
- Follows pointers and interfaces to reach the underlying struct.
- Optional recursion into nested structs via the `RECURSIVE` flag.
- Optional code generation (`GenerateCode`) that emits a `Default()` method with plain assignments, so filling costs nothing at runtime.

## Installation

```sh
go get github.com/firmanmm/go-default
```

The package name is `godefault`, which differs from the last element of the module path, so alias it on import:

```go
import godefault "github.com/firmanmm/go-default"
```

## Usage

### Filling a struct at runtime

Call `Fill` with a pointer to the struct (or the struct value itself) and a flag value. Pass `0` when you don't need recursion, or `godefault.RECURSIVE` to also fill nested structs:

```go
package main

import (
	"fmt"
	"time"

	godefault "github.com/firmanmm/go-default"
)

type Config struct {
	Name        string        `default:"anonymouse"`
	LuckyNumber int           `default:"-13"`
	Age         uint          `default:"23"`
	Height      float64       `default:"180.5"`
	Active      bool          `default:"true"`
	Bytes       []byte        `default:"MTIzNA=="` // base64 -> "1234"
	Strings     []string      `default:"first,second,third"`
	Ints        []int         `default:"-1,0,1"`
	Floats      []float64     `default:"111.111,222.222"`
	Duration    time.Duration `default:"1m10s"`
}

func main() {
	cfg := Config{} // all fields zero
	if err := godefault.Fill(&cfg, 0); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Name)       // anonymouse
	fmt.Println(cfg.LuckyNumber) // -13
	fmt.Println(cfg.Duration)   // 1m10s
}
```

Fields that are already non-zero are left untouched:

```go
cfg := Config{Name: "custom"}
godefault.Fill(&cfg, 0)
// cfg.Name stays "custom"; the other zero fields are filled
```

### Nested structs

By default only the top-level struct is filled. Use the `RECURSIVE` flag to descend into embedded/nested struct fields:

```go
type Inner struct {
	RecurName string `default:"recur-anonymouse"`
}

type Outer struct {
	Name       string `default:"anonymouse"`
	Recurrence Inner
}

func main() {
	var o Outer
	if err := godefault.Fill(&o, godefault.RECURSIVE); err != nil {
		panic(err)
	}
	fmt.Println(o.Recurrence.RecurName) // recur-anonymouse
}
```

## Tag reference

| Tag | Applies to | Meaning | Example |
| --- | --- | --- | --- |
| `default` | any supported field | Default value assigned when the field is zero. Slices/arrays split the value on commas; `[]byte` values are base64-encoded; `time.Duration` values use `time.ParseDuration` syntax. | `` `default:"anonymouse"` `` |
| `default-opt` | any field with `default` | Comma-separated options modifying how the default is applied. The only recognized option is `nonzero`, which is meaningful for generated code: the assignment is wrapped in a zero check (see [Code generation](#code-generation)). Runtime `Fill` already skips non-zero fields unconditionally, so this option has no additional effect there. | `` `default:"23" default-opt:"nonzero"` `` |

Supported field kinds for `default`: `string`, `int`/`int8`/`int16`/`int32`/`int64`, `uint`/`uint8`/`uint16`/`uint32`/`uint64`, `float32`, `float64`, `bool`, slices and arrays of those kinds, `[]byte`, and `time.Duration`. Any other kind (e.g. `complex128`) returns an error.

## Error handling

`Fill` and `GenerateCode` return errors. Predefined sentinel errors:

| Error | When it occurs |
| --- | --- |
| `ErrNilValue` | The passed value is `nil`. |
| `ErrUnsupportedValue` | The (dereferenced) value is not a struct. |
| `ErrUnsupportedOperation` | Reserved sentinel for operations not supported by the library. |

In addition, per-field failures return descriptive errors: an unsupported field kind yields ``Unsupported Value Given for field name "..."``, and malformed default values yield the underlying `strconv`/`time.ParseDuration`/base64 parse error.

## Code generation

Reflection is fine for most applications, but if you want zero runtime cost, `GenerateCode` writes a `Default()` method for your struct to any `io.Writer`, using plain assignments instead of reflection:

```go
package main

import (
	"bytes"
	"fmt"

	godefault "github.com/firmanmm/go-default"
)

type Example struct {
	Name          string `default:"anonymouse"`
	LuckyNumber   int    `default:"-13"`
	SkippedNumber uint   `default:"23" default-opt:"nonzero"`
}

func main() {
	var data Example
	out := bytes.NewBuffer(nil)
	if err := godefault.GenerateCode(&data, out); err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}
```

Output:

```go
func (e *Example) Default() {
	e.Name = "anonymouse"
	e.LuckyNumber = -13
	if e.SkippedNumber == 0 {
		e.SkippedNumber = 23
	}
}
```

Save this output into a `.go` file next to your struct (typically via a `go:generate` directive), and you can call `data.Default()` directly with no reflection. Note that, unlike runtime `Fill`, generated code assigns defaults **unconditionally** unless the field carries `default-opt:"nonzero"`, which wraps the assignment in a zero check as shown above. Code generation also does not recurse into nested structs.

## License

MIT License © 2021 Firman "Rendoru" Maulana. See [LICENSE](LICENSE).
