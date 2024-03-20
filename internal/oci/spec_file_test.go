package oci

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/require"
)

func TestLoadFrom(t *testing.T) {
	testCases := []struct {
		contents []byte
		isError  bool
		spec     *specs.Spec
	}{
		{
			contents: []byte{},
			isError:  true,
		},
		{
			contents: []byte("{}"),
			isError:  false,
			spec:     &specs.Spec{},
		},
	}

	for i, tc := range testCases {
		var spec *specs.Spec
		spec, err := LoadFrom(bytes.NewReader(tc.contents))

		if tc.isError {
			require.Error(t, err, "%d: %v", i, tc)
		} else {
			require.NoError(t, err, "%d: %v", i, tc)
		}

		if tc.spec == nil {
			require.Nil(t, spec, "%d: %v", i, tc)
		} else {
			require.EqualValues(t, tc.spec, spec, "%d: %v", i, tc)
		}
	}
}

func TestFlushTo(t *testing.T) {
	testCases := []struct {
		isError  bool
		spec     *specs.Spec
		contents string
	}{
		{
			spec: nil,
		},
		{
			spec:     &specs.Spec{},
			contents: "{\"ociVersion\":\"\"}\n",
		},
	}

	for i, tc := range testCases {
		buffer := bytes.Buffer{}

		err := flushTo(tc.spec, &buffer)

		if tc.isError {
			require.Error(t, err, "%d: %v", i, tc)
		} else {
			require.NoError(t, err, "%d: %v", i, tc)
		}

		require.EqualValues(t, tc.contents, buffer.String(), "%d: %v", i, tc)
	}

	// Add a simple test for a writer that returns an error when writing
	err := flushTo(&specs.Spec{}, errorWriter{})
	require.Error(t, err)
}

// errorWriter implements the io.Writer interface, always returning an error when
// writing.
type errorWriter struct{}

func (e errorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("error writing")
}
