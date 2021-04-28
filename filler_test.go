package godefault

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFill(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			"Given Struct With Zero Value It Should Return It Filled",
			args{
				&struct {
					Name        string        `default:"anonymouse"`
					LuckyNumber int           `default:"-13"`
					Age         uint          `default:"23"`
					Height      float64       `default:"180.5"`
					Bytes       []byte        `default:"MTIzNA=="`
					Strings     []string      `default:"first,second,third"`
					Ints        []int         `default:"-1,0,1"`
					Uints       []uint        `default:"0,1,2"`
					Floats      []float64     `default:"111.111,222.222,333.333"`
					Duration    time.Duration `default:"1m10s"`
				}{},
			},
			&struct {
				Name        string        `default:"anonymouse"`
				LuckyNumber int           `default:"-13"`
				Age         uint          `default:"23"`
				Height      float64       `default:"180.5"`
				Bytes       []byte        `default:"MTIzNA=="`
				Strings     []string      `default:"first,second,third"`
				Ints        []int         `default:"-1,0,1"`
				Uints       []uint        `default:"0,1,2"`
				Floats      []float64     `default:"111.111,222.222,333.333"`
				Duration    time.Duration `default:"1m10s"`
			}{
				Name:        "anonymouse",
				LuckyNumber: -13,
				Age:         23,
				Height:      180.5,
				Bytes:       []byte{'1', '2', '3', '4'},
				Strings:     []string{"first", "second", "third"},
				Ints:        []int{-1, 0, 1},
				Uints:       []uint{0, 1, 2},
				Floats:      []float64{111.111, 222.222, 333.333},
				Duration:    70 * time.Second,
			},
			false,
		},
		{
			"Given Struct With Partial Zero Value It Should Return Only Non Zero Value",
			args{
				&struct {
					Name        string `default:"anonymouse"`
					NameNonZero string `default:"subzero" default-opt:"nonzero"`
				}{
					NameNonZero: "Nope",
				},
			},
			&struct {
				Name        string `default:"anonymouse"`
				NameNonZero string `default:"subzero" default-opt:"nonzero"`
			}{
				Name:        "anonymouse",
				NameNonZero: "Nope",
			},
			false,
		},
		{
			"Given Struct With Partial Zero Value It Should Return Only Non Zero Value",
			args{
				&struct {
					Name  string `default:"anonymouse"`
					NoTag string
				}{},
			},
			&struct {
				Name  string `default:"anonymouse"`
				NoTag string
			}{
				Name: "anonymouse",
			},
			false,
		},
		{
			"Given Struct With Partial Unexported Value It Should Return Only Exported Value",
			args{
				&struct {
					Name   string `default:"anonymouse"`
					hidden string `default:"subzero"`
				}{},
			},
			&struct {
				Name   string `default:"anonymouse"`
				hidden string `default:"subzero"`
			}{
				Name: "anonymouse",
			},
			false,
		},
		{
			"Given Unsuppoted Field Then It Should Fail",
			args{
				&struct {
					Complex complex128 `default:"123.3"`
				}{},
			},
			nil,
			true,
		},
		{
			"Given Nil Then It Should Fail",
			args{
				nil,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Fill(tt.args.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.args.data)
		})
	}
}
