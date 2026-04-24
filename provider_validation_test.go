package xtreamcodes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// sampleDataRoot resolves the workspace _sample_data/iptv-proxy directory.
// Lookup order:
//  1. XTREAM_SAMPLE_DATA_DIR env var (absolute path)
//  2. Workspace-relative ../../_sample_data/iptv-proxy
//  3. Skip the test if neither resolves
func sampleDataRoot(t *testing.T) string {
	t.Helper()
	if env := os.Getenv("XTREAM_SAMPLE_DATA_DIR"); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env
		}
		t.Fatalf("XTREAM_SAMPLE_DATA_DIR set to %q but path does not exist", env)
	}
	fallback := filepath.Join("..", "..", "_sample_data", "iptv-proxy")
	if _, err := os.Stat(fallback); err == nil {
		abs, _ := filepath.Abs(fallback)
		return abs
	}
	t.Skip("sample data not found at $XTREAM_SAMPLE_DATA_DIR or ../../_sample_data/iptv-proxy")
	return ""
}

// endpointSpec describes how to decode one Xtream sample file against a target struct.
type endpointSpec struct {
	file string
	// decodeAll lenient-unmarshals the full document. Returns the decoded value,
	// the record count (1 for single-object endpoints, len(slice) for arrays),
	// and any unmarshal error.
	decodeAll func([]byte) (interface{}, int, error)
	// elementsOf returns the per-record JSON fragments for arrays. nil for
	// single-object endpoints.
	elementsOf func([]byte) ([]json.RawMessage, error)
	// newElement returns a fresh pointer to the element type for strict decoding.
	newElement func() interface{}
}

func rawArray(b []byte) ([]json.RawMessage, error) {
	var arr []json.RawMessage
	err := json.Unmarshal(b, &arr)
	return arr, err
}

var validationEndpoints = []endpointSpec{
	{
		file: "auth.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v AuthenticationResponse
			err := json.Unmarshal(b, &v)
			return v, 1, err
		},
		newElement: func() interface{} { return new(AuthenticationResponse) },
	},
	{
		file: "get_live_categories.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []Category
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(Category) },
	},
	{
		file: "get_vod_categories.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []Category
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(Category) },
	},
	{
		file: "get_series_categories.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []Category
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(Category) },
	},
	{
		file: "get_live_streams.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []Stream
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(Stream) },
	},
	{
		file: "get_vod_streams.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []Stream
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(Stream) },
	},
	{
		file: "get_series.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			var v []SeriesInfo
			err := json.Unmarshal(b, &v)
			return v, len(v), err
		},
		elementsOf: rawArray,
		newElement: func() interface{} { return new(SeriesInfo) },
	},
	{
		file: "get_series_info.json",
		// Provider samples wrap a single Series in an array; the upstream
		// single-object form is also supported via dispatch on the leading byte.
		decodeAll: func(b []byte) (interface{}, int, error) {
			t := bytes.TrimLeft(b, " \t\r\n")
			if len(t) > 0 && t[0] == '[' {
				var v []Series
				err := json.Unmarshal(b, &v)
				return v, len(v), err
			}
			var v Series
			err := json.Unmarshal(b, &v)
			return v, 1, err
		},
		elementsOf: func(b []byte) ([]json.RawMessage, error) {
			t := bytes.TrimLeft(b, " \t\r\n")
			if len(t) > 0 && t[0] == '[' {
				return rawArray(b)
			}
			return []json.RawMessage{json.RawMessage(b)}, nil
		},
		newElement: func() interface{} { return new(Series) },
	},
	{
		file: "get_vod_info.json",
		decodeAll: func(b []byte) (interface{}, int, error) {
			t := bytes.TrimLeft(b, " \t\r\n")
			if len(t) > 0 && t[0] == '[' {
				var v []VideoOnDemandInfo
				err := json.Unmarshal(b, &v)
				return v, len(v), err
			}
			var v VideoOnDemandInfo
			err := json.Unmarshal(b, &v)
			return v, 1, err
		},
		elementsOf: func(b []byte) ([]json.RawMessage, error) {
			t := bytes.TrimLeft(b, " \t\r\n")
			if len(t) > 0 && t[0] == '[' {
				return rawArray(b)
			}
			return []json.RawMessage{json.RawMessage(b)}, nil
		},
		newElement: func() interface{} { return new(VideoOnDemandInfo) },
	},
}

// strictDecode decodes data into target with DisallowUnknownFields enabled.
func strictDecode(data []byte, target interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(target)
}

// unknownFieldName extracts the field name from json's "unknown field" error.
// Returns "" for any other error.
func unknownFieldName(err error) string {
	if err == nil {
		return ""
	}
	const prefix = "json: unknown field "
	msg := err.Error()
	i := strings.Index(msg, prefix)
	if i < 0 {
		return ""
	}
	return strings.Trim(msg[i+len(prefix):], "\"")
}

// TestProviderSampleValidation cross-validates go.xtream-codes structs against
// real provider sample data captured under _sample_data/iptv-proxy/Provider*/.
//
// For each (provider, endpoint) combination it:
//   1. Lenient-unmarshals the full document and reports the record count.
//   2. Strict-decodes each record (DisallowUnknownFields) and reports any
//      JSON keys present in provider data but absent from the target struct.
//   3. Round-trips the lenient-decoded value through Marshal and reports
//      semantic differences against the original.
//
// Lenient unmarshal failures and marshal failures fail the test; unknown
// fields and round-trip differences are logged for triage.
func TestProviderSampleValidation(t *testing.T) {
	root := sampleDataRoot(t)

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read sample data root %q: %v", root, err)
	}

	var providers []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "Provider") {
			providers = append(providers, e.Name())
		}
	}
	sort.Strings(providers)
	if len(providers) == 0 {
		t.Skipf("no Provider* directories under %s", root)
	}
	t.Logf("validating %d provider(s) from %s", len(providers), root)

	for _, provider := range providers {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			for _, ep := range validationEndpoints {
				ep := ep
				t.Run(ep.file, func(t *testing.T) {
					path := filepath.Join(root, provider, ep.file)
					data, err := os.ReadFile(path)
					if err != nil {
						t.Skipf("missing %s", path)
						return
					}

					decoded, n, err := ep.decodeAll(data)
					if err != nil {
						t.Errorf("lenient decode failed: %v", err)
						return
					}
					t.Logf("decoded %d record(s)", n)

					unknown := map[string]int{}
					strictErrors := 0
					if ep.elementsOf == nil {
						target := ep.newElement()
						if err := strictDecode(data, target); err != nil {
							if name := unknownFieldName(err); name != "" {
								unknown[name]++
							} else {
								strictErrors++
								t.Logf("strict decode error: %v", err)
							}
						}
					} else {
						elems, err := ep.elementsOf(data)
						if err != nil {
							t.Errorf("split into elements: %v", err)
						} else {
							for _, raw := range elems {
								// Repeatedly strict-decode to surface every distinct
								// unknown field, not just the first one.
								remaining := raw
								for {
									target := ep.newElement()
									err := strictDecode(remaining, target)
									if err == nil {
										break
									}
									name := unknownFieldName(err)
									if name == "" {
										strictErrors++
										break
									}
									if _, seen := unknown[name]; !seen {
										unknown[name] = 0
									}
									unknown[name]++
									// Strip the offending key from the raw JSON so the
									// next iteration can find the next unknown field.
									stripped, ok := stripJSONKey(remaining, name)
									if !ok {
										break
									}
									remaining = stripped
								}
							}
						}
					}
					if len(unknown) > 0 {
						names := make([]string, 0, len(unknown))
						for k := range unknown {
							names = append(names, k)
						}
						sort.Strings(names)
						for _, k := range names {
							t.Logf("unknown field %q in %d record(s)", k, unknown[k])
						}
					}
					if strictErrors > 0 {
						t.Logf("non-unknown-field strict decode errors: %d", strictErrors)
					}

					marshalled, err := json.Marshal(decoded)
					if err != nil {
						t.Errorf("marshal failed: %v", err)
						return
					}
					diffs := jsonSemanticDiff(data, marshalled, 50)
					if len(diffs) > 0 {
						t.Logf("round-trip differences: %d (showing first 50)", len(diffs))
						for _, d := range diffs {
							t.Logf("  %s", d)
						}
					}
				})
			}
		})
	}
}

// stripJSONKey removes the first occurrence of a top-level key from a JSON
// object byte slice. Returns the modified bytes and true on success.
// This is used to iteratively reveal multiple unknown fields per record.
func stripJSONKey(data []byte, key string) ([]byte, bool) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	if _, ok := m[key]; !ok {
		return nil, false
	}
	delete(m, key)
	out, err := json.Marshal(m)
	if err != nil {
		return nil, false
	}
	return out, true
}

// jsonSemanticDiff decodes both inputs to interface{} and walks them, returning
// up to limit human-readable difference descriptions. Path syntax: $ for root,
// .key for object descent, [i] for array index. For arrays larger than 5
// elements only the first 3 are walked element-wise; element-uniform drift
// makes deeper walking redundant for diagnostic purposes.
func jsonSemanticDiff(a, b []byte, limit int) []string {
	var av, bv interface{}
	if err := json.Unmarshal(a, &av); err != nil {
		return []string{fmt.Sprintf("decode original: %v", err)}
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return []string{fmt.Sprintf("decode round-trip: %v", err)}
	}
	var diffs []string
	walkDiff("$", av, bv, &diffs, limit)
	return diffs
}

func walkDiff(path string, a, b interface{}, diffs *[]string, limit int) {
	if len(*diffs) >= limit {
		return
	}
	if reflect.DeepEqual(a, b) {
		return
	}
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			*diffs = append(*diffs, fmt.Sprintf("%s: type changed from object to %T", path, b))
			return
		}
		keys := map[string]struct{}{}
		for k := range av {
			keys[k] = struct{}{}
		}
		for k := range bv {
			keys[k] = struct{}{}
		}
		sortedKeys := make([]string, 0, len(keys))
		for k := range keys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		for _, k := range sortedKeys {
			if len(*diffs) >= limit {
				return
			}
			ax, aok := av[k]
			bx, bok := bv[k]
			switch {
			case aok && !bok:
				*diffs = append(*diffs, fmt.Sprintf("%s.%s: present in original (%s), dropped in round-trip", path, k, summary(ax)))
			case !aok && bok:
				*diffs = append(*diffs, fmt.Sprintf("%s.%s: not in original, added in round-trip (%s)", path, k, summary(bx)))
			default:
				walkDiff(path+"."+k, ax, bx, diffs, limit)
			}
		}
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok {
			*diffs = append(*diffs, fmt.Sprintf("%s: type changed from array to %T", path, b))
			return
		}
		if len(av) != len(bv) {
			*diffs = append(*diffs, fmt.Sprintf("%s: array length %d -> %d", path, len(av), len(bv)))
			return
		}
		walkN := len(av)
		if walkN > 3 {
			walkN = 3
		}
		for i := 0; i < walkN; i++ {
			if len(*diffs) >= limit {
				return
			}
			walkDiff(fmt.Sprintf("%s[%d]", path, i), av[i], bv[i], diffs, limit)
		}
	default:
		*diffs = append(*diffs, fmt.Sprintf("%s: %s -> %s", path, summary(a), summary(b)))
	}
}

func summary(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	return s
}
