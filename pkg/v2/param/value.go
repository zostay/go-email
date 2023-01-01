package param

import (
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/zostay/go-email/pkg/email/v2"
)

const (
	Filename = "filename"
	Charset = "charset"
	Boundary = "boundary"
)

type Value struct {
	v string
	ps map[string]string
}

func Parse(v string) (*Value, error) {
	mt, ps, err := mime.ParseMediaType(v)
	if err != nil {
		return nil, err
	}

	return &Value{mt, ps}, nil
}

func New(v string) *Value {
	return &Value{v, map[string]string{}}, nil
}

func NewWithParams(v string, ps map[string]string) *Value {
	return &Value{v, ps}, nil
}

type Modifier func(*Value)

func Change(value string) Modifier {
	return func(pv *Value) {
		pv.v = value
	}
}

func Set(name, value string) Modifier {
	return func(pv *Value) {
		pv.ps[name] = value
	}
}

func Delete(name string) Modifier {
	return func(pv *Value) {
		delete(pv.ps, name)
	}
}

func Modify(pv *Value, changes ...Modifier) *Value {
	copy := pv.Clone()
	for _, change := range changes {
		change(copy)
	}
	return copy
}

func (pv *Value) Value() string {
	return pv.v
}

func (pv *Value) Disposition() string {
	return pv.v
}

func (pv *Value) MediaType() string {
	return pv.v
}

func (pv *Value) Type() string {
	if ix := strings.IndexRune(pv.v, '/'); ix >= 0 {
		return pv.v[:ix]
	}
	return ""
}

func (pv *Value) Subtype() string {
	if ix := strings.IndexRune(pv.v, '/'); ix >= 0 {
		return pv.v[ix+1:]
	}
	return ""
}

// func (pv *Value) SetValue(v string) {
// 	pv.v = v
// }

func (pv *Value) Parameters() map[string]string {
	return pv.ps
}

func (pv *Value) Parameter(k string) string {
	return pv.ps[k]
}

// func (pv *Value) SetParameter(k, v string) {
// 	if v == "" {
// 		delete(pv.ps, k)
// 		return
// 	}
// 	pv.ps[k] = v
// }

func (pv *Value) Filename() string {
	return pv.ps[Filename]
}

// func (pv *Value) SetFilename(f string) {
// 	pv.SetParameter(cdFilename, f)
// }

func (pv *Value) Charset() string {
	return pv.ps[Charset]
}

// func (pv *Value) SetCharset(c string) {
// 	pv.SetParameter(ctCharset, c)
// }

func (pv *Value) Boundary() string {
	return pv.ps[Boundary]
}

// func (pv *Value) SetBoundary(b string) {
// 	pv.SetParameter(ctBoundary, b)
// }

func (pv *Value) String() string {
	pks := make([]string, 0, len(pv.ps))
	for k := range pv.ps {
		pks = append(pks, k)
	}
	sort.Strings(pks)

	parts := make([]string, len(pv.ps)+1)
	parts[0] = pv.v

	for n, k := range pks {
		parts[n+1] = fmt.Sprintf("%s=%s", k, pv.ps[k])
	}

	return strings.Join(parts, "; ")
}

func (pv *Value) Bytes() []byte {
	return []byte(pv.String())
}

func (pv *Value) Clone() *Value {
	var copy Value
	copy.v= pv.v
	copy.ps = make(map[string]string, len(pv.ps))
	for k, v := range pv.ps {
		copy.ps[k] = v
	}
	return &copy
}

var _ email.Outputter = &Value{}
