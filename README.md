<p align="center">
    <h3 align="center">GoMud</h3>
    <p align="center">Go 语言写的，支持 UTF-8 的中文 MUD 客户端</p>
    <p align="center">
        <a href="#如何使用-gomud">如何使用</a> •
        <a href="#安装指南">如何安装</a> •
        <a href="#配置-gomud">如何配置</a> •
        <a href="https://github.com/mudclient/go-mud/wiki/FAQ">常见问题</a>
    </p>
    <p align="center">
<a href="https://github.com/mudclient/go-mud/releases/latest">
<img alt="最新版本" src="https://img.shields.io/github/v/release/mudclient/go-mud.svg?logo=github&style=flat-square">
</a>
<a href="https://github.com/mudclient/go-mud/actions?workflow=Release">
<img alt="Release workflow" src="https://github.com/mudclient/go-mud/workflows/Release/badge.svg">
</a>
<a href="https://github.com/mudclient/go-mud/actions?workflow=Build">
<img alt="Build workflow" src="https://github.com/mudclient/go-mud/workflows/Build/badge.svg">
</a>
<a href="https://goreportcard.com/report/github.com/mudclient/go-mud">
<img alt="Go Report" src="https://goreportcard.com/badge/github.com/mudclient/go-mud">
</a>
<a href="https://github.com/pre-commit/pre-commit">
<img alt="pre-commit" src="https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit">
</a>
    </p>
</p>

本项目实现目标是一个 MUD 客户端，主要采用 Go 语言实现。

本项目基于中国知名 MUD 游戏《[北大侠客行](http://www.pkuxkx.com)》开发，但也应该可以适用于其它 MUD 服务。

## 什么是 MUD

MUD（/mʌd/, 参见[维基百科](https://zh.wikipedia.org/zh-cn/MUD)），原指多用户地牢（Multi-User Dungeon），
通常将缩写直译为“网络泥巴”或是简称“泥巴”（英文 **mud** 的意思为泥巴）。

MUD 是一种多人即时的虚拟世界，通常以文字描述为基础。其中结合了**角色扮演**、**江湖**、**互动小说**与**在线聊天**等元素，玩家可以阅读或查看房间、物品、其他玩家、非玩家角色的描述，并在虚拟世界中做特定动作。玩家通常会透过输入类似自然语言的指令（如`drink`, `eat`, `bow`）来与虚拟世界中的物品或其他玩家交互。

MUD 一般认为是最早的网络游戏，历史悠久，内涵丰富。在古朴的终端界面下，通过阅读文字展开想象，来构筑一个庞大的虚拟世界，因此富有独特的魅力。

## 什么是北大侠客行

北大侠客行（以下称**北侠**）于 1996 年开服，至今仍在运营，算是国内运行非常长的网络游戏了。
而且这些年一直都有更新，实属难能可贵。

基于 MUD 特有的文化，挂机在北侠也是被允许的，而在 MUD 下开发挂机程序也是一种别有风味的玩法。

## GoMud 有什么用

GoMud 是一个 MUD 客户端，可以用来连接 MUD 服务器，提供纯文本的用户界面，以供玩家与 MUD 服务器交互。

**GoMud 目前仍在开发当中，并不完善。** 但已经能够提供必要的功能来连接 MUD 服务器。且有许多亮点：

* [X] 全程使用 UTF-8，天生免疫乱码
* [X] 支持运行有 GB2312/GBK/GB18030/BIG5/UTF-16/UTF-8 编码的 MUD 服务器
* [X] 纯文本界面，通过命令行和快捷键来操作
* [X] 支持 macOS、Linux、Windows、安卓四大平台
* [X] 支持**路由器**、**树莓派**、**群晖**、**电视机**等小众平台
* [X] 支持 32 位和 64 位操作系统
* [X] 支持 [Lua 机器人](https://github.com/dzpao/lua-mud-robots)

## GoMud 不能做什么

* 不支持图形界面，没有丰富的代码编辑框或诸如此类的其它 UI 元素来帮助你写触发和机器人
* 没有庞大的用户群，没有大量开箱即用的机器人，对伸手党不友好
* 不能帮助没有丰富的计算机操作经验，特别是 *nix 命令行操作经验的人熟悉 *nix
* 不能帮助没有编程经验或者没有学习过 Go 语言的人学会编程、学会 Go 语言

## 如何使用 GoMud

### 运行环境

* GoMud 可在 Linux、macOS 及 Windows 上运行。运行时不依赖其它软件。
* 通过 [Termux](https://termux.com/) 的帮助，GoMud 也可以在安卓下运行。你可以在运行安卓系统的手机或者电视机上使用 GoMud。
* GoMud 也可以在群晖 NAS、运行有 OpenWRT 等 Linux 系统的智能路由器，或者树莓派上运行。

### 安装指南

本项目的[发布页面](https://github.com/mudclient/go-mud/releases)
中包含了所有支持平台的预编译安装包，你可以根据自己的需要选择下载。

#### macOS 快速安装

macOS 用户推荐使用 [Homebrew](https://brew.sh) 来安装。如果你没用过它，不如趁此机会安装体验一下。

```sh
brew tap mudclient/tap
brew install go-mud
```

#### Termux 快速安装

运行了安卓系统的手机、平板电脑、电视机通过 Termux 也可以使用 GoMud，安装方法如下：

```sh
wget https://github.com/mudclient/go-mud/releases/download/v0.6.1/go-mud_v0.6.1_Termux_ARMv7.deb
apt install ./go-mud_v0.6.1_Termux_ARMv7.deb
```

以上命令以 ARMv7 架构上 v0.6.1 版本的为例，
其它版本及架构请前往发布页面选择相应的预编译安装包。
如果你不知道自己设备的 CPU 架构，可以通过 `uname -m` 命令获知。

#### 手动安装

本项目的发布页面中包含了所有支持平台的预编译可执行文件。
各平台的可执行文件名称略有不同，你可以下载和你的运行环境相对应的版本。
GoMud 支持的平台非常丰富，限于篇幅，此处不再赘述。
更多内容请查看[支持平台与安装指南](https://github.com/mudclient/go-mud/wiki/支持平台与安装指南)。

#### 通过源码安装

GoMud 采用 Go 语言实现，如果你要通过源码安装，则需要自行准备 Golang 开发环境。
推荐使用 Go 1.13 或以上的版本。Golang 安装完毕后，通过如下命令序列即可安装：

```
git clone https://github.com/mudclient/go-mud.git
cd go-mud
go generate ./...
go build
```

### 启动并进入北侠

下述示例中的程序文件名假定为 `go-mud`，如果你采用的是预编译的可执行文件，
你可能需要下载后改名或者将下述命令中的程序文件名替换为真实的程序文件名称。

```
$ go-mud
初始化 Lua 环境...
Lua 环境初始化完成。
连接到服务器 mud.pkuxkx.net:8080...连接成功。
...
```

### 配置 GoMud

GoMud 支持通过配置文件或者命令行选项的方式来指定程序运行参数，
目前已有的命令行选项如下：

```
$ go-mud --help       # 可以获得使用帮助
GoMud(version v0.6.1)

Usage:
  go-mud [flags]

Flags:
  -c, --config FILENAME            config FILENAME, default to `config.yaml` or `config.json`
      --version                    just print version number only
  -h, --help                       show this message
      --gen-yaml                   generate config.yaml
      --gen-json                   generate config.json
      --ui.ambiguouswidth string   二义性字符宽度，可选值: auto/single/double/space (default "auto")
      --ui.historylines int        历史记录保留行数 (default 100000)
      --ui.rttvheight int          历史查看模式下实时文本区域高度 (default 10)
  -H, --mud.host IP/Domain         服务器 IP/Domain (default "mud.pkuxkx.net")
  -P, --mud.port Port              服务器 Port (default 8080)
      --mud.encoding Encoding      服务器输出文本的 Encoding (default "UTF-8")
      --lua.enable                 是否加载 Lua 机器人 (default true)
  -p, --lua.path path              Lua 插件路径 path (default "lua")
```

配置文件同时支持 [YAML](https://yaml.org/) 和 [JSON](https://json.org/) 两种格式，
两种配置文件效果是一样的，用户可根据个人偏好选择使用，下面分别给出示例。
配置项的说明参见[配置与运行](https://github.com/mudclient/go-mud/wiki/配置与运行)。

#### config.yaml 示例

默认的 YAML 配置文件名为 `config.yaml`，如果省略配置文件，等同默认内容如下：

```yaml
UI:
  AmbiguousWidth: auto
  HistoryLines: 100000
  RTTVHeight: 10
MUD:
  Host: mud.pkuxkx.net
  Port: 8080
  Encoding: UTF-8
Lua:
  Enable: true
  Path: lua
```

#### config.json 示例

默认的 JSON 配置文件名为 `config.json`，如果省略配置文件，等同默认内容如下：

```json
{
  "UI": {
    "AmbiguousWidth": "auto",
    "HistoryLines": 100000,
    "RTTVHeight": 10
  },
  "Mud": {
    "Host": "mud.pkuxkx.net",
    "Port": 8080,
    "Encoding": "UTF-8"
  },
  "Lua": {
    "Enable": true,
    "Path": "lua"
  }
}
```

### 通过 Docker 来启动

GoMud 也可支持通过 Docker 来运行，推荐使用 Docker 来挂机。

```
待完善
```

## 如何贡献

* 体验并向周围的人分享你的体验结果
* 通过[提交 issue](https://github.com/mudclient/go-mud/issues/new) 来反馈意见
* 通过 PR 来贡献代码，贡献代码时请先阅读[贡献指南](CONTRIBUTING.md)
