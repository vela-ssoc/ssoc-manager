package mresponse

type BinarySupport struct {
	Name          string     `json:"name"`          // 显示名字
	Value         string     `json:"value"`         // 值
	Architectures NameValues `json:"architectures"` // 支持的架构
}

type BinarySupports []*BinarySupport

func DefaultBinarySupports() BinarySupports {
	x64 := &NameValue{Name: "x64", Value: "amd64"}
	x86 := &NameValue{Name: "x86", Value: "386"}
	arm64 := &NameValue{Name: "ARM64", Value: "arm64"}
	arm32 := &NameValue{Name: "ARM32", Value: "arm"}
	loong64 := &NameValue{Name: "LoongArch64", Value: "loong64"}
	riscv64 := &NameValue{Name: "RISC-V 64-bit", Value: "riscv64"}

	return []*BinarySupport{
		{
			Name:          "Linux",
			Value:         "linux",
			Architectures: NameValues{x64, x86, arm64, arm32, loong64, riscv64},
		},
		{
			Name:  "Windows",
			Value: "windows",
			Architectures: NameValues{
				x64, x86, arm64, arm32,
			},
		},
		{
			Name:          "macOS",
			Value:         "darwin",
			Architectures: NameValues{x64, arm64},
		},
	}
}
