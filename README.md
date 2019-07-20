# go-mud —— Go 语言写的 MUD 客户端

本项目实现目标是一个 MUD 客户端，主要采用 Go 语言实现。

本项目基于中国知名 MUD 游戏《[北大侠客行](http://www.pkuxkx.com)》开发，但也应该可以适用于其它 MUD 服务。

## 如何使用本程序

### 运行环境

+ 本程序可在 Linux 及 macOS 上运行。运行时不依赖其它软件。
+ 建议golang版本 >= 1.11

### 启动并进入北侠
+ Linux
```
  > echo 'export GO111MODULE=on'>>/etc/profile
  > echo 'export GOPROXY=https://goproxy.io'>>/etc/profile
  > source /etc/profile
  > cd $GOPATH/src
  > git clone https://github.com/dzpao/go-mud.git
  > cd go-mud
  > git submodule update --init
  > go run main.go
```

+ Windows
```
待完善
```

## 本程序有什么用

**本程序目前仍在开发当中，并不完善。** 但已经能够提供必要的功能来登录 MUD 服务器。且有许多亮点：

* [X] 原生支持 UTF-8
* [X] 纯文本界面，通过命令行和快捷键来操作
* [X] 支持 macOS、Linux、Windows、安卓四大平台
* [X] 支持 32 位和 64 位操作系统
* [X] 支持 [Lua 机器人](https://github.com/dzpao/lua-mud-robots)

## 本程序不能做什么

* 不支持图形界面，没有丰富的代码编辑框或诸如此类的其它 UI 元素来帮助你写触发和机器人
* 没有庞大的用户群，没有大量开箱即用的机器人，对伸手党不友好

## 如何使用本程序

### 运行环境

本程序可在 Linux 及 macOS 上运行。运行时不依赖其它软件。

### 启动并进入北侠

```
待完善
```

### 通过 Docker 来启动

本项目支持通过 Docker 来运行，推荐使用 Docker 来挂机。

```
待完善
```

## 如何贡献

* 体验
* 通过[提交 issue](https://github.com/dzpao/go-mud/issues/new) 来反馈意见
* 通过 PR 来贡献代码

## 什么是 MUD

MUD（/mʌd/, 参见[维基百科](https://zh.wikipedia.org/zh-cn/MUD)），原指多用户地牢（Multi-User Dungeon），
通常将缩写直译为“网络泥巴”或是简称“泥巴”（英文 **mud** 的意思为泥巴）。

MUD 是一种多人即时的虚拟世界，通常以文字描述为基础。其中结合了**角色扮演**、**江湖**、**互动小说**与**在线聊天**等元素，玩家可以阅读或查看房间、物品、其他玩家、非玩家角色的描述，并在虚拟世界中做特定动作。玩家通常会透过输入类似自然语言的指令（如`drink`, `eat`, `bow`）来与虚拟世界中的物品或其他玩家交互。

MUD 一般认为是最早的网络游戏，历史悠久，内涵丰富。在古朴的终端界面下，通过阅读文字展开想象，来构筑一个庞大的虚拟世界，因此富有独特的魅力。

## 什么是北大侠客行

北大侠客行（以下称 **北侠**）于 1996 年开服，至今仍在运营，算是国内运行非常长的网络游戏了。
而且这些年一直都有更新，实属难能可贵。

基于 MUD 特有的文化，挂机在北侠也是被允许的，而在 MUD 下开发挂机程序也是一种别有风味的玩法。
