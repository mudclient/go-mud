package app

import (
	"fmt"
)

//go:generate go run gen.go

func VersionDetail() string {
	info := fmt.Sprintf("版本信息如下：\n程序版本: %s\n编译时间: %s\n编译环境: %s\n编译设备: %s\n其中：\n",
		Version, BuildTime, BuildGoVersion, BuildHost)
	for _, contributor := range Contributors {
		info += fmt.Sprintf("    %s 贡献了 %d 行代码\n", contributor.Name, contributor.Lines)
	}

	return info
}
