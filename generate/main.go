package main

// TODO ref against core/src/main/scala/kafka/server/KafkaApis.scala
// ~line 111 main handler for all requests
// Add CanVersion or something to check whether features are being used
// on an unsupported version.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

var maxKey int

func die(why string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, why+"\n", args...)
	os.Exit(1)
}

type (
	// LineWriter writes lines at a time.
	LineWriter struct {
		buf bytes.Buffer
	}

	Type interface {
		WriteAppend(*LineWriter)
		WriteDecode(*LineWriter)
		TypeName() string
	}

	Bool           struct{}
	Int8           struct{}
	Int16          struct{}
	Int32          struct{}
	Int64          struct{}
	Uint32         struct{}
	Varint         struct{}
	Varlong        struct{}
	String         struct{}
	NullableString struct{}
	Bytes          struct{}
	NullableBytes  struct{}
	VarintString   struct{}
	VarintBytes    struct{}

	Array struct {
		Inner         Type
		IsVarintArray bool
	}

	StructField struct {
		Comment    string
		MinVersion int
		FieldName  string
		Type       Type
	}

	Struct struct {
		TopLevel bool
		Comment  string
		Name     string

		Admin        bool   // only relevant if TopLevel
		Key          int    // only relevant if TopLevel
		MinVersion   int    // only relevant if TopLevel
		MaxVersion   int    // only relevant if TopLevel
		ResponseKind string // only relevant if TopLevel

		Fields []StructField
	}
)

func (l *LineWriter) Write(line string, args ...interface{}) {
	fmt.Fprintf(&l.buf, line, args...)
	l.buf.WriteByte('\n')
}

//go:generate sh -c "go run . | gofmt > ../kmsg/messages.go"
func main() {
	f, err := ioutil.ReadFile("MESSAGES")
	if err != nil {
		die("unable to read MESSAGES file: %v", err)
	}
	Parse(f)

	l := new(LineWriter)
	l.Write("package kmsg")
	l.Write(`import "github.com/twmb/kgo/kbin"`)
	l.Write("// Code generated by kgo/generate. DO NOT EDIT.\n")

	l.Write("const MaxKey = %d\n", maxKey)

	for _, s := range newStructs {
		s.WriteDefn(l)
		if s.TopLevel {
			if s.ResponseKind != "" {
				s.WriteKeyFunc(l)
				s.WriteMaxVersionFunc(l)
				s.WriteMinVersionFunc(l)
				s.WriteSetVersionFunc(l)
				s.WriteGetVersionFunc(l)
				if s.Admin {
					s.WriteAdminFunc(l)
				}
				s.WriteResponseKindFunc(l)
				l.Write("") // newline before append func
				s.WriteAppendFunc(l)
			} else {
				s.WriteDecodeFunc(l)
			}
		}
	}

	fmt.Println(l.buf.String())
}