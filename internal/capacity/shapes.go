package capacity

var KnownShapes = []string{
	"BM.DenseIO1.36", "BM.DenseIO2.52", "BM.DenseIO.E4.128", "BM.DenseIO.E5.128", "BM.GPU2.2", "BM.GPU3.8", "BM.GPU4.8", "BM.GPU.A10.4", "BM.GPU.A100-v2.8",
	"BM.GPU.H100.8", "BM.GPU.L40S.4", "BM.GPU.MI300X.8", "BM.GPU.MI355X.8", "BM.GPU.H200.8", "BM.GPU.B200.8", "BM.GPU.GB200.4", "BM.GPU.GB300.4", "BM.HPC2.36",
	"BM.HPC.E5.144", "BM.Optimized3.36", "BM.Standard1.36", "BM.Standard2.52", "BM.Standard3.64", "BM.Standard.A1.160", "BM.Standard.A4.48", "BM.Standard.B1.44",
	"BM.Standard.E2.64", "BM.Standard.E3.128", "BM.Standard.E4.128", "BM.Standard.E5.192", "BM.Standard.E6.256", "VM.DenseIO1.16", "VM.DenseIO1.4", "VM.DenseIO1.8",
	"VM.DenseIO2.16", "VM.DenseIO2.24", "VM.DenseIO2.8", "VM.DenseIO.E4.Flex", "VM.DenseIO.E5.Flex", "VM.GPU2.1", "VM.GPU3.1", "VM.GPU3.2", "VM.GPU3.4", "VM.GPU.A10.1",
	"VM.GPU.A10.2", "VM.Optimized3.Flex", "VM.Standard1.1", "VM.Standard1.16", "VM.Standard1.2", "VM.Standard1.4", "VM.Standard1.8", "VM.Standard2.1", "VM.Standard2.16",
	"VM.Standard2.2", "VM.Standard2.24", "VM.Standard2.4", "VM.Standard2.8", "VM.Standard3.Flex", "VM.Standard.A1.Flex", "VM.Standard.A2.Flex", "VM.Standard.A4.Flex",
	"VM.Standard.B1.1", "VM.Standard.B1.16", "VM.Standard.B1.2", "VM.Standard.B1.4", "VM.Standard.B1.8", "VM.Standard.E2.1", "VM.Standard.E2.1.Micro", "VM.Standard.E2.2",
	"VM.Standard.E2.4", "VM.Standard.E2.8", "VM.Standard.E3.Flex", "VM.Standard.E4.Flex", "VM.Standard.E5.Flex", "VM.Standard.E6.Flex",
}

var DenseIOFlexShapes = map[string][]string{
	"VM.DenseIO.E4.Flex": {"8", "16", "32"},
	"VM.DenseIO.E5.Flex": {"8", "16", "24", "32", "40", "48"},
}

type DenseIOConfig struct {
	OCPUs       float32
	MemoryInGBs float32
	NVMes       int
}

var DenseIOShapeConfigs = map[string]map[float32]DenseIOConfig{
	"VM.DenseIO.E4.Flex": {
		8:  {OCPUs: 8, MemoryInGBs: 128, NVMes: 1},
		16: {OCPUs: 16, MemoryInGBs: 256, NVMes: 2},
		32: {OCPUs: 32, MemoryInGBs: 512, NVMes: 4},
	},
	"VM.DenseIO.E5.Flex": {
		8:  {OCPUs: 8, MemoryInGBs: 96, NVMes: 1},
		16: {OCPUs: 16, MemoryInGBs: 192, NVMes: 2},
		24: {OCPUs: 24, MemoryInGBs: 288, NVMes: 3},
		32: {OCPUs: 32, MemoryInGBs: 384, NVMes: 4},
		40: {OCPUs: 40, MemoryInGBs: 480, NVMes: 5},
		48: {OCPUs: 48, MemoryInGBs: 576, NVMes: 6},
	},
}