package palette

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const minInt64 int64 = -1 << 63

var (
	errEmptyCdf = errors.New("can not create a CDF for a empty histogram")
)

// Histogram is simple histogram implementation.
type Histogram []bin

// Cdf is a cumulative distribution function (CDF) of colors in a image.
// The value field is the less than or equal to the probability of this color.
type Cdf []bin

type bin struct {
	Key   int     // The color.
	Value float64 // The value.
}

func (x bin) String() string {
	return fmt.Sprintf("(%d: %f)", x.Key, x.Value)
}

// MarshalBinary encodes the receiver into a binary form and returns the result.
func (h Histogram) MarshalBinary() (data []byte, err error) {
	sort.Slice(h, func(i, j int) bool {
		return h[i].Key < h[j].Key
	})

	var out BinaryEncoder

	// Number of bins
	if err := out.Varint(int64(len(h))); err != nil {
		return nil, fmt.Errorf("error writing length: %s", err)
	}

	prev := minInt64
	for _, x := range h {
		key := int64(x.Key)
		diff := uint64(key - prev) // This is in sorted order, so x.Key should always be larger than prev.
		prev = key

		if err := out.Uvarint(diff); err != nil {
			return nil, fmt.Errorf("error writing key: %s", err)
		}
		if err := out.Float64(x.Value); err != nil {
			return nil, fmt.Errorf("error writing value: %s", err)
		}
	}

	return out.Bytes(), nil
}

func UnMarshalBinaryHistogram(data []byte) (Histogram, error) {
	b := bytes.NewReader(data)

	length, err := binary.ReadVarint(b)
	if err != nil {
		return nil, fmt.Errorf("error reading length: %s", err)
	}

	bins := make([]bin, int(length))
	prev := minInt64
	for i := 0; i < len(bins); i++ {
		diff, err := binary.ReadUvarint(b)
		if err != nil {
			return nil, fmt.Errorf("error reading key: %s", err)
		}

		var value float64
		if err := binary.Read(b, binary.LittleEndian, &value); err != nil {
			return nil, fmt.Errorf("error reading value: %s", err)
		}

		prev += int64(diff)
		bins[i] = bin{
			Key:   int(prev),
			Value: value,
		}
	}

	return Histogram(bins), nil
}

// Cdf returns the cumulative distribution function (CDF) for this histogram.
func (h Histogram) Cdf() (Cdf, error) {
	if len(h) == 0 {
		return nil, errEmptyCdf
	}

	// Ensure the histogram is normalised.
	cdf := Cdf(h.Normalise())

	// Sort by the probability of each color.
	sort.Slice(cdf, func(i, j int) bool {
		if cdf[i].Value == cdf[j].Value {
			return cdf[i].Key < cdf[j].Key
		}
		return cdf[i].Value < cdf[j].Value
	})

	// Convert the value to a running total.
	var total float64
	for i := range cdf {
		cdf[i].Value += total
		total = cdf[i].Value
	}

	return cdf, nil
}

// Combine combines two histograms together. The resulting histogram is
// not normalised.
func (h Histogram) Combine(hist Histogram) Histogram {
	m := make(map[int]float64)
	for _, x := range h {
		m[x.Key] += x.Value
	}
	for _, x := range hist {
		m[x.Key] += x.Value
	}

	return mapToHistogram(m, 1)
}

// Normalise ensures the area under the histogram equals one.
func (h Histogram) Normalise() Histogram {
	total := h.Total() // TODO Test for total being zero.

	hist := make([]bin, len(h))
	for i, x := range h {
		hist[i].Key = x.Key
		hist[i].Value = x.Value / total
	}

	return hist
}

func (h Histogram) String() string {
	return fmt.Sprintf("Histogram(Len: %d, Total: %f)", len(h), h.Total())
}

// Total returns the area under the histogram.
func (h Histogram) Total() float64 {
	var total float64
	for _, x := range h {
		total += x.Value
	}
	return total
}

func mapToHistogram(m map[int]float64, factor float64) Histogram {
	hist := make([]bin, 0, len(m))
	for key, count := range m {
		hist = append(hist, bin{Key: key, Value: count / factor})
	}
	return Histogram(hist)
}

// Rand is a source of random float64s, compatible with *rand.Rand.
type Rand interface {
	// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0)
	Float64() float64
}

// Rand returns a random value from the cdf, chosen based on the weight of the values.
func (cdf Cdf) Rand(random Rand) int {
	if len(cdf) == 0 {
		return 0
	}

	p := random.Float64()
	i := sort.Search(len(cdf), func(i int) bool {
		return cdf[i].Value >= p
	})
	return cdf[i].Key
}
