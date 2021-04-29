package godefault

import (
	"bytes"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestGenerateCode(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := GenerateCode(tt.args.data, out); (err != nil) != tt.wantErr {
				t.Errorf("GenerateCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("GenerateCode() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func ExampleGenerateCode() {

	type Example struct {
		Name          string        `default:"anonymouse"`
		LuckyNumber   int           `default:"-13"`
		Age           uint          `default:"23"`
		Height        float64       `default:"180.5"`
		Bytes         []byte        `default:"MTIzNA=="`
		Strings       []string      `default:"first,second,third"`
		Ints          []int         `default:"-1,0,1"`
		Uints         []uint        `default:"0,1,2"`
		Floats        []float64     `default:"111.111,222.222,333.333"`
		Duration      time.Duration `default:"1m10s"`
		SkippedNumber uint          `default:"23" default-opt:"nonzero"`
		SkippedString string        `default:"Original" default-opt:"nonzero"`
		SkippedArray  []float64     `default:"9,8,7" default-opt:"nonzero"`
		SkippedBytes  []byte        `default:"SGVsbG8gV29ybGQhIEhhdmUgQSBHcmVhdCBEYXkgOj4=" default-opt:"nonzero"`
	}

	data := Example{}
	writer := bytes.NewBuffer(nil)
	if err := GenerateCode(&data, writer); err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(writer.String())
	// Output:
	// func (e *Example) Default() {
	// 	e.Name = "anonymouse"
	// 	e.LuckyNumber = -13
	// 	e.Age = 23
	// 	e.Height = 180.500000
	// 	e.Bytes = []byte {
	// 		0x31, 0x32, 0x33, 0x34,
	// 	}
	// 	e.Strings = []string {
	// 		"first",
	// 		"second",
	// 		"third",
	// 	}
	// 	e.Ints = []int {
	// 		-1,
	// 		0,
	// 		1,
	// 	}
	// 	e.Uints = []uint {
	// 		0,
	// 		1,
	// 		2,
	// 	}
	// 	e.Floats = []float64 {
	// 		111.111000,
	// 		222.222000,
	// 		333.333000,
	// 	}
	// 	e.Duration = time.Duration(70000000000)
	// 	if e.SkippedNumber == 0 {
	// 		e.SkippedNumber = 23
	// 	}
	// 	if e.SkippedString == "" {
	// 		e.SkippedString = "Original"
	// 	}
	// 	if e.SkippedArray == nil {
	// 		e.SkippedArray = []float64 {
	// 			9.000000,
	// 			8.000000,
	// 			7.000000,
	// 		}
	// 	}
	// 	if e.SkippedBytes == nil {
	// 		e.SkippedBytes = []byte {
	// 			0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c,
	// 			0x64, 0x21, 0x20, 0x48, 0x61, 0x76, 0x65, 0x20, 0x41, 0x20,
	// 			0x47, 0x72, 0x65, 0x61, 0x74, 0x20, 0x44, 0x61, 0x79, 0x20,
	// 			0x3a, 0x3e,
	// 		}
	// 	}
	// }
}
