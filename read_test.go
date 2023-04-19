package streamtools

import (
	"strings"
	"testing"
)

type readUntilTest struct {
	Input     string
	Byte      byte
	Output    string
	Completed bool
}

func TestReadUntil(t *testing.T) {
	for idx, test := range []readUntilTest{
		{
			"abcd\nefgh",
			10,
			"abcd",
			true,
		},
		{
			"0123456789\nabc",
			10,
			"0123456789",
			false,
		},
		{
			"0123456789\n",
			10,
			"0123456789",
			false,
		},
		{
			"p=78&x=moo",
			'&',
			"p=78",
			true,
		},
	} {
		r := strings.NewReader(test.Input)
		buf := make([]byte, 10)
		n, done, err := ReadUntil(r, test.Byte, buf)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if n != len(test.Output) {
			t.Fatalf("in test %d, wrong length: %q",
				idx, string(buf[:n]))
		}
		if done != test.Completed {
			t.Fatalf("in test %d, expected done to be %v, but it was %v",
				idx, test.Completed, done)
		}

		if string(buf[:n]) != test.Output {
			t.Fatalf("in test %d, got %q but expected %q",
				idx, string(buf[:n]), test.Output)
		}
	}
}

type readUntilAnyTest struct {
	Input     string
	Bytes     []byte
	Output    string
	Completed bool
}

func TestReadUntilAny(t *testing.T) {
	for idx, test := range []readUntilAnyTest{
		{
			"abcd\nefgh",
			[]byte{10, 1},
			"abcd",
			true,
		},
		{
			"abcd\nefgh",
			[]byte{1, 10},
			"abcd",
			true,
		},
		{
			"0123456789\nabc",
			[]byte{10},
			"0123456789",
			false,
		},
		{
			"0123456789\n",
			[]byte{10},
			"0123456789",
			false,
		},
		{
			"p=78&x=moo",
			[]byte{'&'},
			"p=78",
			true,
		},
		{
			"012345678",
			nil,
			"012345678",
			true,
		},
		{
			"0123456789",
			nil,
			"0123456789",
			false,
		},
	} {
		r := strings.NewReader(test.Input)
		buf := make([]byte, 10)
		n, done, err := ReadUntilAny(r, test.Bytes, buf)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if n != len(test.Output) {
			t.Fatalf("in test %d, wrong length: %q",
				idx, string(buf[:n]))
		}
		if done != test.Completed {
			t.Fatalf("in test %d, expected done to be %v, but it was %v",
				idx, test.Completed, done)
		}

		if string(buf[:n]) != test.Output {
			t.Fatalf("in test %d, got %q but expected %q",
				idx, string(buf[:n]), test.Output)
		}
	}
}
