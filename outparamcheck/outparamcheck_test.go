// Copyright 2013 Kamil Kisiel
// Modifications copyright 2016 Palantir Technologies, Inc.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package outparamcheck

import (
	"go/token"
	"io/ioutil"
	"path"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestOutParamCheck(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected []OutParamError
	}{
		{
			name: "interface",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				j := []byte("...")
				var x interface{}
				json.Unmarshal(j, x)
				json.Unmarshal(j, &x)
				json.Unmarshal(j, *&x)
				json.Unmarshal(j, nil)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   146,
						Line:     11,
						Column:   23,
					},
					Line:     `json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "assignment",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				j := []byte("...")
				var x interface{}
				_ = json.Unmarshal(j, x)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   150,
						Line:     11,
						Column:   27,
					},
					Line:     `_ = json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "go",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				j := []byte("...")
				var x interface{}
				go json.Unmarshal(j, x)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   149,
						Line:     11,
						Column:   26,
					},
					Line:     `go json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "defer",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				j := []byte("...")
				var x interface{}
				defer json.Unmarshal(j, x)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   152,
						Line:     11,
						Column:   29,
					},
					Line:     `defer json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "send",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				c := make(chan error)
				j := []byte("...")
				var x interface{}
				c <- json.Unmarshal(j, x)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   177,
						Line:     12,
						Column:   28,
					},
					Line:     `c <- json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "return",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func foo() error {
				j := []byte("...")
				var x interface{}
				return json.Unmarshal(j, x)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   158,
						Line:     11,
						Column:   30,
					},
					Line:     `return json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "in case statement",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			func main() {
				j := []byte("...")
				var x interface{}
				switch {
					case json.Unmarshal(j, x) == nil:
						_ = x
				}
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   165,
						Line:     12,
						Column:   29,
					},
					Line:     `case json.Unmarshal(j, x) == nil:`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "in struct literal assignment",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			type errStruct struct {
				err error
			}

			func main() {
				j := []byte("...")
				var x interface{}
				_ = errStruct {
					err: json.Unmarshal(j, x), 
				}
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   219,
						Line:     16,
						Column:   29,
					},
					Line:     `err: json.Unmarshal(j, x),`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "struct based",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			type  A struct{}

			func main() {		
				x := A{}
				pointerX := &x
				
				j := []byte("...")
				json.Unmarshal(j, x)
				json.Unmarshal(j, pointerX)
				json.Unmarshal(j, &x)
				json.Unmarshal(j, *&x)
				json.Unmarshal(j, nil)
			}
			`,
			expected: []OutParamError{
				{
					Pos: token.Position{
						Filename: "", // will be filled in by the test case run
						Offset:   184,
						Line:     15,
						Column:   23,
					},
					Line:     `json.Unmarshal(j, x)`,
					Method:   "Unmarshal",
					Argument: 1,
				},
			},
		},
		{
			name: "declared within if block",
			input: `
			package main
			
			import (
				"encoding/json"
			)
			
			type  A struct{}

			func main() {		
				x := A{}
				j := []byte("...")
				
				if pointerX := &x; true{
					json.Unmarshal(j, pointerX)
				}
			}
			`,
		},
	}

	tmpDir, cleanup, err := dirs.TempDir(".", "")
	require.NoError(t, err)
	defer cleanup()

	for _, tc := range tcs {
		// write program to temp file
		currCaseDir, err := ioutil.TempDir(tmpDir, "")
		require.NoError(t, err)

		fpath := path.Join(currCaseDir, "main.go")
		err = ioutil.WriteFile(fpath, []byte(tc.input), 0644)
		require.NoError(t, err)

		// load package for program
		pkgs, err := packages.Load(&packages.Config{
			Mode: packages.LoadSyntax,
		}, "./"+currCaseDir)
		require.NoError(t, err)

		// update the expected outparam output filename
		for i := range tc.expected {
			tc.expected[i].Pos.Filename = pkgs[0].GoFiles[0]
		}

		// run out-param checker
		errs := run(pkgs, defaultCfg)

		// assert expectations
		assert.Equal(t, tc.expected, errs)
	}
}
