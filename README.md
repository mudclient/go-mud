# go-mud —— Go 语言写的 MUD 客户端

本项目实现目标是一个 MUD 客户端，主要采用 Go 语言实现。

本项目基于中国知名 MUD 游戏《[北大侠客行](http://www.pkuxkx.com)》开发，但也应该可以适用于其它 MUD 服务。

## 什么是 MUD

MUD（/mʌd/, 参见[维基百科](https://zh.wikipedia.org/zh-cn/MUD)），原指多用户地牢（Multi-User Dungeon），
通常将缩写直译为“网络泥巴”或是简称“泥巴”（英文 **mud** 的意思为泥巴）。

MUD 是一种多人即时的虚拟世界，通常以文字描述为基础。其中结合了**角色扮演**、**江湖**、**互动小说**与**在线聊天**等元素，玩家可以阅读或查看房间、物品、其他玩家、非玩家角色的描述，并在虚拟世界中做特定动作。玩家通常会透过输入类似自然语言的指令（如`drink`, `eat`, `bow`）来与虚拟世界中的物品或其他玩家交互。

MUD 一般认为是最早的网络游戏，历史悠久，内涵丰富。在古朴的终端界面下，通过阅读文字展开想象，来构筑一个庞大的虚拟世界，因此富有独特的魅力。

## 什么是北大侠客行

北大侠客行（以下称 **北侠**）于 1996 年开服，至今仍在运营，算是国内运行非常长的网络游戏了。
而且这些年一直都有更新，实属难能可贵。

基于 MUD 特有的文化，挂机在北侠也是被允许的，而在 MUD 下开发挂机程序也是一种别有风味的玩法。

## GoMud 有什么用

GoMud 是一个 MUD 客户端，可以用来连接 MUD 服务器，提供纯文本的用户界面，以供玩家与 MUD 服务器交互。

**GoMud 目前仍在开发当中，并不完善。** 但已经能够提供必要的功能来连接 MUD 服务器。且有许多亮点：

* [X] 原生支持 UTF-8
* [X] 纯文本界面，通过命令行和快捷键来操作
* [X] 支持 macOS、Linux、Windows、安卓四大平台
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
* 通过 [Termux](https://termux.com/) 的帮助，GoMud 也可以在安卓下运行。

### 通过源码安装

GoMud 采用 Go 语言实现，如果你要通过源码安装，则需要自行准备 Golang 开发环境。
Golang 安装完毕后，通过如下 `go get` 命令即可安装：

```
go get -u github.com/dzpao/go-mud
```

### 下载预编译的可执行文件

本项目的 [Release](https://github.com/dzpao/go-mud/releases) 中包含了所有支持平台的预编译可执行文件。
各平台的可执行文件名称略有不同，统一格式为 `go-mud-<系统名称>-<硬件架构>`，
你可以下载和你的运行环境相对应的版本，直接开始使用。

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
go-mud(version 0.6)

Usage:
  go-mud [flags]

Flags:
  -c, --config FILENAME            config FILENAME, default to `config.yaml` or `config.json`
      --version                    just print version number only
      --help                       show this message
      --gen-yaml                   generate config.yaml
      --gen-json                   generate config.json
      --ui.ambiguouswidth string   二义性字符宽度 (default "auto")
  -H, --mud.host IP/Domain         服务器 IP/Domain (default "mud.pkuxkx.net")
  -P, --mud.port Port              服务器 Port (default 8080)
      --lua.enable                 是否加载 Lua 机器人 (default true)
  -p, --lua.path path              Lua 插件路径 path (default "lua")
```

配置文件同时支持 [YAML](https://yaml.org/) 和 [JSON](https://json.org/) 两种格式，
两种配置文件效果是一样的，用户可根据个人偏好选择使用，下面分别给出示例。
配置项的说明参见 [Wiki](https://github.com/dzpao/go-mud/wiki/configuration)。

#### config.yaml 示例

默认的 YAML 配置文件名为 `config.yaml`，如果省略配置文件，等同默认内容如下：

```yaml
UI:
  AmbiguousWidth: auto
MUD:
  Host: mud.pkuxkx.net
  Port: 8080
Lua:
  Enable: true
  Path: lua
```

#### config.json 示例

默认的 JSON 配置文件名为 `config.json`，如果省略配置文件，等同默认内容如下：

```json
{
    "UI": {
        "AmbiguousWidth": "auto"
    },
    "Mud": {
        "Host": "mud.pkuxkx.net",
        "Port": 8080
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
* 通过[提交 issue](https://github.com/dzpao/go-mud/issues/new) 来反馈意见
* 通过 PR 来贡献代码，贡献代码时请先阅读[贡献指南](CONTRIBUTING.md)
