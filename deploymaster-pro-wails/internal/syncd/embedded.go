package syncd

import "embed"

//go:embed bin/syncd_linux_amd64 bin/syncd_linux_arm64 bin/syncd_darwin_amd64 bin/syncd_darwin_arm64
var binaries embed.FS

func GetLinuxAMD64() []byte {
	data, _ := binaries.ReadFile("bin/syncd_linux_amd64")
	return data
}

func GetLinuxARM64() []byte {
	data, _ := binaries.ReadFile("bin/syncd_linux_arm64")
	return data
}

func GetDarwinAMD64() []byte {
	data, _ := binaries.ReadFile("bin/syncd_darwin_amd64")
	return data
}

func GetDarwinARM64() []byte {
	data, _ := binaries.ReadFile("bin/syncd_darwin_arm64")
	return data
}
