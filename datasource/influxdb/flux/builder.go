package flux

import (
	"fmt"
	"reflect"
)

type keyWord string
type Operator string

const (
	from        keyWord = "from"
	r           keyWord = "range"
	filt        keyWord = "filter"
	pipeForward keyWord = "|>"
	bucket      keyWord = "bucket"
	start       keyWord = "start"
	end         keyWord = "end"
)

const (
	Addition       Operator = "+"
	Subtraction    Operator = "-"
	Multiplication Operator = "*"
	Division       Operator = "/"
	Exponentiation Operator = "^"
	Modulo         Operator = "%"
	Equal          Operator = "=="
	NotEqual       Operator = "!="
	Less           Operator = "<"
	Greater        Operator = ">"
	LessEqual      Operator = "<="
	GreaterEqual   Operator = ">="
)

type Builder struct {
	// b bucket
	b string
	// r range
	r duration
	// f filters
	f []filter
}

type duration struct {
	start string
	end   string
}

type filter struct {
	field string
	op    Operator
	value interface{}
	vType reflect.Type
}

func New() *Builder {
	return &Builder{}
}

func (b *Builder) Bucket(bucket string) *Builder {
	b.b = bucket
	return b
}

func (b *Builder) Range(start string, end string) *Builder {
	b.r = duration{start: start, end: end}
	return b
}

func (b *Builder) AddFilter(field string, op Operator, value interface{}) *Builder {
	b.f = append(b.f, filter{
		field: field,
		op:    op,
		value: value,
		vType: reflect.TypeOf(value),
	})
	return b
}

func (b *Builder) Build() string {
	var res string
	// from(bucket:"bucket")
	res += fmt.Sprintf("%s(%s:\"%s\")", from, bucket, b.b)
	res += fmt.Sprintf(" %s %s(%s:%s", pipeForward, r, start, b.r.start)
	// |> range(start:-1h)
	if b.r.end == "" {
		res += fmt.Sprintf(")")
	} else {
		// |> range(start:-1h,end:-5m)
		res += fmt.Sprintf(", %s:%s)", end, b.r.end)
	}
	// |> filter(fn: (r) =>
	res += fmt.Sprintf(" %s %s(fn: (r) =>", pipeForward, filt)
	count := len(b.f)
	for _, v := range b.f {
		if v.vType.Kind() == reflect.String {
			res += fmt.Sprintf(" r.%s %s \"%v\"", v.field, v.op, v.value)
		} else {
			res += fmt.Sprintf("r.%s %s %v", v.field, v.op, v.value)
		}
		count--
		if count != 0 {
			res += " and"
		}
	}
	res += fmt.Sprintf(" )")
	return res
}
