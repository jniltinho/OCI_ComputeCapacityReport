package capacity

import (
	"testing"

	"github.com/oracle/oci-go-sdk/v65/core"
)

func TestGetShapeConfigDenseIOFlex(t *testing.T) {
	cfg := GetShapeConfig("VM.DenseIO.E4.Flex", nil, 16, 0)

	if !cfg.IsFlex {
		t.Fatal("expected flex shape")
	}
	if cfg.OCPUs != 16 {
		t.Fatalf("expected 16 ocpus, got %v", cfg.OCPUs)
	}
	if cfg.Memory != 0 {
		t.Fatalf("expected 0 memory for denseio flex, got %v", cfg.Memory)
	}
}

func TestGetShapeConfigFlexClampsMemory(t *testing.T) {
	maxOcpus := float32(64)
	maxMem := float32(1024)
	maxPerOcpu := float32(64)

	shapes := []core.Shape{{
		Shape: strPtr("VM.Standard.E5.Flex"),
		OcpuOptions: &core.ShapeOcpuOptions{
			Max: &maxOcpus,
		},
		MemoryOptions: &core.ShapeMemoryOptions{
			MaxInGBs:        &maxMem,
			MaxPerOcpuInGBs: &maxPerOcpu,
		},
	}}

	cfg := GetShapeConfig("VM.Standard.E5.Flex", shapes, 10, 2)

	if cfg.OCPUs != 10 {
		t.Fatalf("expected 10 ocpus, got %v", cfg.OCPUs)
	}
	if cfg.Memory != 10 {
		t.Fatalf("expected memory clamped to 10, got %v", cfg.Memory)
	}
}

func TestFormatValue(t *testing.T) {
	cases := map[float64]string{
		-1:   "-",
		0:    "0",
		4:    "4",
		8.5:  "8.5",
	}

	for input, want := range cases {
		if got := formatValue(input); got != want {
			t.Fatalf("formatValue(%v) = %q, want %q", input, got, want)
		}
	}
}

func TestDenseIOShapeConfigs(t *testing.T) {
	cfg, ok := DenseIOShapeConfigs["VM.DenseIO.E5.Flex"][32]
	if !ok {
		t.Fatal("expected denseio e5 32 ocpus config")
	}
	if cfg.OCPUs != 32 || cfg.MemoryInGBs != 384 || cfg.NVMes != 4 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func strPtr(s string) *string { return &s }