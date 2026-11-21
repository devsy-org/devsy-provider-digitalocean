package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

const testErrorLevel = "error"

func TestProviderLoggingLevels(t *testing.T) {
	for _, tc := range []struct {
		level string
		want  string
	}{
		{"", "info,warn,error"},
		{testErrorLevel, testErrorLevel},
		{"warn", "warn,error"},
		{"info", "info,warn,error"},
		{"debug", "debug,info,warn,error"},
		{"trace", "trace,debug,info,warn,error"},
		{"TrAcE", "trace,debug,info,warn,error"},
	} {
		t.Run("level="+tc.level, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", tc.level)
			t.Setenv("DEVSY_LOG_FORMAT", "auto")
			t.Setenv("TOKEN", "digitalocean-secret-canary")
			output := captureProviderDiagnostics(t, func() {
				log.Trace("trace")
				log.Debug("debug")
				log.Info("info")
				log.Warn("warn")
				log.ErrorContext(
					context.Background(),
					"error digitalocean-secret-canary",
					log.AttrProvider("digitalocean"),
				)
			})
			if bytes.Contains(output, []byte("digitalocean-secret-canary")) {
				t.Fatal("diagnostics contain the provider token")
			}
			levels := diagnosticLevels(t, output)
			if got := strings.Join(levels, ","); got != tc.want {
				t.Fatalf("levels = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInvalidLoggingConfigurationStopsCommand(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"DEVSY_LOG_LEVEL", "invalid-secret-canary"},
		{"DEVSY_LOG_LEVEL", " info"},
		{"DEVSY_LOG_FORMAT", "invalid-secret-canary"},
		{"DEVSY_LOG_FORMAT", " json"},
	} {
		t.Run(tc.name+"="+tc.value, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "")
			t.Setenv("DEVSY_LOG_FORMAT", "")
			t.Setenv(tc.name, tc.value)
			ran := false
			root := NewRootCmd()
			root.AddCommand(
				&cobra.Command{Use: "probe", Run: func(*cobra.Command, []string) { ran = true }},
			)
			root.SetArgs([]string{"probe"})
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.name) ||
				strings.Contains(err.Error(), tc.value) {
				t.Fatalf("unexpected configuration error: %v", err)
			}
			if ran {
				t.Fatal("command ran with invalid logging configuration")
			}
		})
	}
}

func TestSupportedLogFormats(t *testing.T) {
	for _, format := range []string{"", "auto", "AUTO", "json", "JSON", "text", "plain", "PLAIN", "logfmt", "LOGFMT"} {
		t.Run("format="+format, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "info")
			t.Setenv("DEVSY_LOG_FORMAT", format)
			output := captureProviderDiagnostics(t, func() { log.Info("format probe") })
			isJSON := json.Valid(bytes.TrimSpace(output))
			if wantJSON := format == "" || strings.EqualFold(format, "auto") ||
				strings.EqualFold(format, "json"); isJSON != wantJSON {
				t.Fatalf("JSON output = %v, want %v: %s", isJSON, wantJSON, output)
			}
			if !bytes.Contains(output, []byte("format probe")) {
				t.Fatal("diagnostic message missing")
			}
		})
	}
}

func diagnosticLevels(t *testing.T, output []byte) []string {
	t.Helper()
	var levels []string
	for line := range bytes.SplitSeq(bytes.TrimSpace(output), []byte("\n")) {
		var record struct {
			Time     string `json:"ts"`
			Level    string `json:"level"`
			Message  string `json:"msg"`
			Provider string `json:"provider"`
		}
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("invalid JSON diagnostic: %v", err)
		}
		if record.Time == "" || record.Message == "" {
			t.Fatal("diagnostic missing timestamp or message")
		}
		if record.Level == testErrorLevel &&
			(record.Provider != "digitalocean" || !strings.Contains(record.Message, testErrorLevel)) {
			t.Fatal("error diagnostic lost its message or attribute")
		}
		levels = append(levels, record.Level)
	}
	return levels
}

func captureProviderDiagnostics(t *testing.T, emit func()) []byte {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stderr
	os.Stderr = writer
	defer func() { os.Stderr = previous }()

	if err := initProviderLogging(); err != nil {
		t.Fatal(err)
	}
	emit()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return output
}
