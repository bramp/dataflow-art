package palette

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestRGBA_MarshalText(t *testing.T) {
	tests := []struct {
		c    RGBA
		want string
	}{
		{c: RGBA{0, 0, 0, 255}, want: "#000000"},
		{c: RGBA{255, 0, 0, 255}, want: "#ff0000"},
		{c: RGBA{0, 255, 0, 255}, want: "#00ff00"},
		{c: RGBA{0, 0, 255, 255}, want: "#0000ff"},
		{c: RGBA{0, 0, 0, 0}, want: "#000000"},
		{c: RGBA{255, 255, 255, 255}, want: "#ffffff"},
	}

	for _, test := range tests {
		got, err := test.c.MarshalText()
		if err != nil {
			t.Errorf("%v.MarshalText() err %s, want nil", test.c, err)
		}
		if string(got) != test.want {
			t.Errorf("%v.MarshalText() = %q, want %q", test.c, got, test.want)
		}
	}

}

func TestRGBA_UnmarshalText(t *testing.T) {
	tests := []struct {
		hex  string
		want RGBA
	}{
		{hex: "#000000", want: RGBA{0, 0, 0, 255}},
		{hex: "#ff0000", want: RGBA{255, 0, 0, 255}},
		{hex: "#00ff00", want: RGBA{0, 255, 0, 255}},
		{hex: "#0000ff", want: RGBA{0, 0, 255, 255}},
		{hex: "#ffffff", want: RGBA{255, 255, 255, 255}},
	}

	for _, test := range tests {
		var got RGBA
		if err := got.UnmarshalText([]byte(test.hex)); err != nil {
			t.Errorf("UnmarshalText(%q) err %s, want nil", test.hex, err)
		}
		if diff := pretty.Compare(got, test.want); diff != "" {
			t.Errorf("UnmarshalText(%s): diff: (-got +want)\n%s", test.hex, diff)
		}
	}
}
