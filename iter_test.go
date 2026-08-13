package radix

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"testing/quick"
)

var mixedLenKeys = []string{
	"barbazboo", "f", "foo", "found", "zap", "zip",
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIterator_SeekLowerBound(t *testing.T) {
	cases := []struct {
		keys   []string
		search string
		want   []string
	}{
		{mixedLenKeys, "b", []string{"barbazboo", "f", "foo", "found", "zap", "zip"}},
		{mixedLenKeys, "bar", []string{"barbazboo", "f", "foo", "found", "zap", "zip"}},
		{mixedLenKeys, "barbazboo0", []string{"f", "foo", "found", "zap", "zip"}},
		{mixedLenKeys, "zi", []string{"zip"}},
		{mixedLenKeys, "zippy", nil},
		{[]string{"bar", "baz", "foo", "foobar", "qux"}, "fooq", []string{"qux"}},
		{[]string{"f", "fo", "foo", "food", "bug"}, "foo", []string{"foo", "food"}},
		{[]string{"bar", "foo00", "foo11"}, "foo", []string{"foo00", "foo11"}},
		{[]string{"bb", "bc"}, "ac", []string{"bb", "bc"}},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("case%03d", idx), func(t *testing.T) {
			r := New()
			for _, k := range tc.keys {
				if _, updated := r.Insert(k, nil); updated {
					t.Fatalf("duplicate key %q", k)
				}
			}

			iter := r.Iterator()
			iter.SeekLowerBound(tc.search)

			var out []string
			for {
				k, _, ok := iter.Next()
				if !ok {
					break
				}
				out = append(out, k)
			}
			if !stringSlicesEqual(out, tc.want) {
				t.Fatalf("search=%q got=%v want=%v", tc.search, out, tc.want)
			}
		})
	}
}

func TestIterator_SeekLowerBound_BinaryKeys(t *testing.T) {
	keys := []string{
		string([]byte{0, 0, 0, 1}),
		string([]byte{0, 0, 0, 2}),
		string([]byte{0, 0, 0, 10}),
		string([]byte{0, 0, 1, 0}),
	}
	r := New()
	for i, k := range keys {
		r.Insert(k, i)
	}

	seek := string([]byte{0, 0, 0, 2})
	iter := r.Iterator()
	iter.SeekLowerBound(seek)

	k, v, ok := iter.Next()
	if !ok || k != seek || v.(int) != 1 {
		t.Fatalf("expected key at index 1, got k=%q v=%v ok=%v", k, v, ok)
	}
}

func TestIterator_SeekPrefix(t *testing.T) {
	r := New()
	for _, k := range []string{"foo", "foobar", "foobaz", "bar"} {
		r.Insert(k, k)
	}

	iter := r.Iterator()
	iter.SeekPrefix("foo")

	var out []string
	for {
		k, _, ok := iter.Next()
		if !ok {
			break
		}
		out = append(out, k)
	}
	want := []string{"foo", "foobar", "foobaz"}
	if !stringSlicesEqual(out, want) {
		t.Fatalf("got=%v want=%v", out, want)
	}
}

type readableString string

func (s readableString) Generate(rand *rand.Rand, size int) reflect.Value {
	const letters = "abcdefg"
	size = rand.Intn(8)
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return reflect.ValueOf(readableString(b))
}

func TestIterator_SeekLowerBound_Fuzz(t *testing.T) {
	r := New()
	var set []string

	radixAddAndScan := func(newKey, searchKey readableString) []string {
		r.Insert(string(newKey), nil)

		it := r.Iterator()
		var result []string
		it.SeekLowerBound(string(searchKey))
		for {
			k, _, ok := it.Next()
			if !ok {
				break
			}
			result = append(result, k)
		}
		return result
	}

	sliceAddSortAndFilter := func(newKey, searchKey readableString) []string {
		set = append(set, string(newKey))
		sort.Strings(set)

		var result []string
		for i, k := range set {
			if i > 0 && set[i-1] == k {
				continue
			}
			if k >= string(searchKey) {
				result = append(result, k)
			}
		}
		return result
	}

	if err := quick.CheckEqual(radixAddAndScan, sliceAddSortAndFilter, nil); err != nil {
		t.Fatal(err)
	}
}
