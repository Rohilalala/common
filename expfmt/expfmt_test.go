// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expfmt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/common/model"
)

// Test Format to Escaping Scheme conversion.
func TestToFormatType(t *testing.T) {
	tests := []struct {
		format   Format
		expected FormatType
	}{
		{
			format:   FmtProtoCompact,
			expected: TypeProtoCompact,
		},
		{
			format:   FmtProtoDelim,
			expected: TypeProtoDelim,
		},
		{
			format:   FmtProtoText,
			expected: TypeProtoText,
		},
		{
			format:   FmtOpenMetrics_1_0_0,
			expected: TypeOpenMetrics,
		},
		{
			format:   FmtText,
			expected: TypeTextPlain,
		},
		{
			format:   FmtOpenMetrics_0_0_1,
			expected: TypeOpenMetrics,
		},
		{
			format:   "application/vnd.google.protobuf; proto=BadProtocol; encoding=text",
			expected: TypeUnknown,
		},
		{
			format:   "application/vnd.google.protobuf",
			expected: TypeUnknown,
		},
		{
			format:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily=bad",
			expected: TypeUnknown,
		},
		// encoding missing
		{
			format:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily",
			expected: TypeUnknown,
		},
		// invalid encoding
		{
			format:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=textual",
			expected: TypeUnknown,
		},
		// bad charset, must be utf-8
		{
			format:   "application/openmetrics-text; version=1.0.0; charset=ascii",
			expected: TypeUnknown,
		},
		{
			format:   "text/plain",
			expected: TypeTextPlain,
		},
		{
			format:   "text/plain; version=invalid",
			expected: TypeUnknown,
		},
		{
			format:   "gobbledygook",
			expected: TypeUnknown,
		},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.format.FormatType())
	}
}

func TestToEscapingScheme(t *testing.T) {
	tests := []struct {
		format   Format
		expected model.EscapingScheme
	}{
		{
			format:   FmtProtoCompact,
			expected: model.UnderscoreEscaping,
		},
		{
			format:   "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=dots",
			expected: model.DotsEscaping,
		},
		{
			format:   "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=allow-utf-8",
			expected: model.NoEscaping,
		},
		// error returns default
		{
			format:   "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=invalid",
			expected: model.NameEscapingScheme,
		},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.format.ToEscapingScheme())
	}
}

func TestWithEscapingScheme(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		scheme   model.EscapingScheme
		expected string
	}{
		{
			name:     "no existing term, append one",
			format:   FmtProtoCompact,
			scheme:   model.DotsEscaping,
			expected: "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=compact-text; escaping=dots",
		},
		{
			name:     "existing term at end, replace",
			format:   "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=underscores",
			scheme:   model.ValueEncodingEscaping,
			expected: "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=values",
		},
		{
			name:     "existing term in middle, replace",
			format:   "application/openmetrics-text; escaping=dots; version=1.0.0; charset=utf-8; ",
			scheme:   model.NoEscaping,
			expected: "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=allow-utf-8",
		},
		{
			name:     "multiple existing terms removed",
			format:   "application/openmetrics-text; escaping=dots; version=1.0.0; charset=utf-8; escaping=allow-utf-8",
			scheme:   model.ValueEncodingEscaping,
			expected: "application/openmetrics-text; version=1.0.0; charset=utf-8; escaping=values",
		},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, string(test.format.WithEscapingScheme(test.scheme)))
	}
}

func TestFormat_Version(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{format: FmtText, expected: "0.0.4"},
		{format: FmtText.WithEscapingScheme(model.NoEscaping), expected: "0.0.4"},
		{format: FmtOpenMetrics_0_0_1, expected: "0.0.1"},
		{format: FmtOpenMetrics_1_0_0, expected: "1.0.0"},
		{format: FmtOpenMetrics_2_0_0, expected: "2.0.0"},
		{format: FmtOpenMetrics_2_0_0.WithEscapingScheme(model.DotsEscaping), expected: "2.0.0"},
		{format: FmtProtoDelim, expected: ""},
		{format: FmtProtoText, expected: ""},
		{format: FmtProtoCompact, expected: ""},
		{format: FmtUnknown, expected: ""},
		{format: Format("invalid"), expected: ""},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, test.format.Version())
	}
}

func TestFormat_Matches(t *testing.T) {
	tests := []struct {
		name     string
		f        Format
		target   Format
		expected bool
	}{
		{
			name:     "FmtText matches FmtText",
			f:        FmtText,
			target:   FmtText,
			expected: true,
		},
		{
			name:     "FmtText matches FmtText with escaping scheme",
			f:        FmtText.WithEscapingScheme(model.NoEscaping),
			target:   FmtText,
			expected: true,
		},
		{
			name:     "FmtText matches FmtText with another escaping scheme",
			f:        FmtText.WithEscapingScheme(model.DotsEscaping),
			target:   FmtText.WithEscapingScheme(model.NoEscaping),
			expected: true,
		},
		{
			name:     "FmtText matches without charset parameter",
			f:        Format("text/plain; version=0.0.4"),
			target:   FmtText,
			expected: true,
		},
		{
			name:     "FmtText matches with whitespace variation",
			f:        Format("text/plain;version=0.0.4;charset=utf-8"),
			target:   FmtText,
			expected: true,
		},
		{
			name:     "FmtText does not match different version",
			f:        Format("text/plain; version=0.0.5; charset=utf-8"),
			target:   FmtText,
			expected: false,
		},
		{
			name:     "FmtText does not match FmtProtoDelim",
			f:        FmtText,
			target:   FmtProtoDelim,
			expected: false,
		},
		{
			name:     "FmtText does not match FmtUnknown",
			f:        FmtText,
			target:   FmtUnknown,
			expected: false,
		},
		{
			name:     "FmtOpenMetrics_1_0_0 matches itself",
			f:        FmtOpenMetrics_1_0_0,
			target:   FmtOpenMetrics_1_0_0,
			expected: true,
		},
		{
			name:     "FmtOpenMetrics_1_0_0 matches with escaping scheme",
			f:        FmtOpenMetrics_1_0_0.WithEscapingScheme(model.ValueEncodingEscaping),
			target:   FmtOpenMetrics_1_0_0,
			expected: true,
		},
		{
			name:     "FmtOpenMetrics_1_0_0 does not match FmtOpenMetrics_2_0_0",
			f:        FmtOpenMetrics_1_0_0,
			target:   FmtOpenMetrics_2_0_0,
			expected: false,
		},
		{
			name:     "FmtOpenMetrics_1_0_0 does not match FmtOpenMetrics_0_0_1",
			f:        FmtOpenMetrics_1_0_0,
			target:   FmtOpenMetrics_0_0_1,
			expected: false,
		},
		{
			name:     "FmtOpenMetrics_2_0_0 matches itself",
			f:        FmtOpenMetrics_2_0_0,
			target:   FmtOpenMetrics_2_0_0,
			expected: true,
		},
		{
			name:     "FmtOpenMetrics_2_0_0 matches with escaping scheme",
			f:        FmtOpenMetrics_2_0_0.WithEscapingScheme(model.NoEscaping),
			target:   FmtOpenMetrics_2_0_0,
			expected: true,
		},
		{
			name:     "FmtProtoDelim matches itself",
			f:        FmtProtoDelim,
			target:   FmtProtoDelim,
			expected: true,
		},
		{
			name:     "FmtProtoDelim matches with escaping scheme",
			f:        FmtProtoDelim.WithEscapingScheme(model.UnderscoreEscaping),
			target:   FmtProtoDelim,
			expected: true,
		},
		{
			name:     "FmtProtoDelim does not match FmtProtoText",
			f:        FmtProtoDelim,
			target:   FmtProtoText,
			expected: false,
		},
		{
			name:     "FmtProtoDelim does not match FmtProtoCompact",
			f:        FmtProtoDelim,
			target:   FmtProtoCompact,
			expected: false,
		},
		{
			name:     "FmtUnknown matches FmtUnknown",
			f:        FmtUnknown,
			target:   FmtUnknown,
			expected: true,
		},
		{
			name:     "FmtUnknown does not match FmtProtoDelim",
			f:        FmtUnknown,
			target:   FmtProtoDelim,
			expected: false,
		},
		{
			name:     "invalid format matches itself",
			f:        Format("not a media type"),
			target:   Format("not a media type"),
			expected: true,
		},
		{
			name:     "invalid format does not match different invalid format",
			f:        Format("not a media type 1"),
			target:   Format("not a media type 2"),
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.f.Matches(test.target))
		})
	}
}
