package palette

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestHistogram_MarshalBinary(t *testing.T) {
	tests := []struct {
		hist Histogram
	}{
		{
			hist: []bin{},
		},
		{
			hist: []bin{{0, 1.0}},
		},
		{
			hist: []bin{{0, 0.5}, {1, 0.5}},
		},
		{
			hist: []bin{{0, 1.0}, {1, 1.0}},
		},
		{
			hist: []bin{{-10, 1.0}, {0, 1.0}, {20, 1.0}, {40, 1.0}},
		},
	}

	for _, test := range tests {
		want := Histogram(test.hist)
		data, err := want.MarshalBinary()
		if err != nil {
			t.Errorf("%q.MarshalBinary() err %s, want nil", test.hist, err)
		}

		got, err := UnMarshalBinaryHistogram(data)
		if err != nil {
			t.Errorf("UnMarshalBinaryHistogram(%q) err %s, want nil", test.hist, err)
		}
		if diff := pretty.Compare(sortBins(got), sortBins(want)); diff != "" {
			t.Errorf("UnMarshalBinaryHistogram(%q): diff: (-got +want)\n%s", test.hist, diff)
		}
	}
}
