# ggbot
一个用Go写的微信机器人

## 缘起
今年(2017)年1月份我表哥找我说，能不能搞一个小工具，可以模拟人工给一个特定的微信公众号回复一些特定的内容，我很爽快的答应了下来。然后通过调研发现用`web`微信的接口可以完成这个需求。在开发的过程中遇到了以下几个问题

* 用什么语言开发呢？由于我用的是`mac os x`， 我表哥是`windows 10`，所以跨平台是必须的 (也正是如此我钟爱的`swift`无缘此次角逐～) `java`、`python`、 `golang` 三选一(当然还有其他很多选择比如世界上最好的语言`PHP`)，`python`需要自己安装运行环境，需要自己装依赖包，对非程序员简直是灾难，排除之～ `java`和`golang`都以编译成一个可执行程序（估计python也可以）但是java 需要依赖 `jre／jdk` 对于非程序员也不友好所以就 是你啦`golang`

* 没用过`golang`啊怎么办， 没事就当练手了～ 所以各位看官如果觉得哪里设计的有问题，或者用法有问题 请一定不吝赐教

* 整个助手小工具做好以后我发现下面几个不舒服的地方

 1. 每次登录都要扫码好蛋疼(一场掉线可以用cookie恢复)
 2. 接口不完整，发个卡片消息都不行
 3. 想抢个红包都做不到
 4. 什么？ 你还想看朋友圈？
 5. 想唯一标识每一个好友？ 做你的春秋大梦吧～
 6. 我不想再吐槽了

那么怎么办呢？***逆向微信*** 然后给微信插入`木马`,让微信成为我们的傀儡，随时接受我们机器人的调遣
*哼哼嘿哈*
由于本人做iOS出身，就从iOS端的微信出发 花了几天时间可以随时从服务端唤醒手机`app`(_好像可以做一些不好的事情，比如给你男朋友装个修改过的微信？_) 然后把小工具代码重构一下摇身一变成为一个机器人 并且他有一个很酷炫的名字 `GGBot`
先看几张图，看个小视频 比较粗糙大家忍耐一下

![cui 截图](http://7xnowv.com1.z0.glb.clouddn.com/Screen-cui.png)

[后台抢红包演示](http://7xnowv.com1.z0.glb.clouddn.com/ggbot-demo.mp4)

## 关于iOS端Tweak(插件)
为了让这个机器人存活的稍微久一点所以我暂时不打算开源这部分东西，如果有需求可以联系我～

## 基本使用

### 如果你不是开发者
可以直接下载我编译好的软件包，通过配置文件就可以运行不同的机器人啦，具体如何配置请继续往下看哦 `windows`，`Mac os x`，`Linux`。如果这个玩具不能满足你可以在开一个`issue`提出你的需求，会有好心的开发者为你实现的。如果你想用这个做生意，坦白来说web微信的接口不适合做生意，因为我认为他稳定性不够好 但是没关系你可以试试看，有哪里不能满足你，你可以开一个`issue`说说看， 说不定有人会帮你呢 程序员 可都是好人哦。当然我强烈建议你用业余时间学学软件开发 ^_^

### 如果你是开发者
我正在考虑写一个提供rpc接口的模块出来，大家多给点意见呀。 如果你是`golang`使用者请一定多看看代码，一起让这个机器人长的壮一点

#### 如何配置安装包

一张图胜过千言万语 把 `ggbot` 拖进记得浏览器然后 `--help`你就会看到下面的提示

![看这里](http://7xnowv.com1.z0.glb.clouddn.com/useage.png)

#### 如何使用源码

我是用go写的，所以你要安装go的开发环境，如果你不想安装那就下载编译好的机器人吧～

[go安装传送门](https://golang.org/doc/install)

安装`wechat`包

```go
go get -u -v github.com/KevinGong2013/ggbot/wechat
```

```go
//main.go

import "github.com/KevinGong2013/ggbot/wechat"

wxbot, _ := wechat.WeakUp(nil)

```

好啦，只要一行代码你就成功的运行起来了一个机器人，怎么样简单吧？
但是这个机器人没有任何功能，仅仅是维持和微信服务器的连接，那么我们要怎么添加新的功能上去呢？很简单

比如我们要打印出来所有收到的消息

```go
// terminal
go get github.com/KevinGong2013/ggbot/modules/echo

// main.go
import "github.com/KevinGong2013/ggbot/modules/echo"

wxbot.RegisterModule(new(echo.Echo))
```

比如我们想要一个和上面的截图类似的 cui 界面

```go
// terminal
go get github.com/KevinGong2013/ggbot/modules/ui

// main.go
import "github.com/KevinGong2013/ggbot/modules/ui"

ui := ui.NewUI(`.ggbot/media`)
wxbot.RegisterModule(ui)
ui.Loop() // 注意ui.Loop() 会阻塞调用
```

## 系统架构

这个标题好宏大 ( ̀⌄ ́) 其实就是想说说这个机器人的内部架构。在接着往下阅读之前，我推荐您先看看这几篇文章
    微信协议
     xxxxxx
     xxxx
好了， 到这里我默认你直到如果通过`wxweb api`和微信`server`沟通.

wechat包，一共有一下几个功能点

    1. 登录 处理所有和登录相关的事情 支持自定义处理UUID
    2. 同步 处理和服务器的同步协议，如果有新事件（联系人变化或者新消息）推送到module去处理
    3. Modules 这是一个很粗糙的模块管理组件，用来分发微信服务器消息
    4. 消息和通讯录部分 这两部分主要是提供一些便利的方法来发送消息 处理通讯录变化 支持自定义消息的发送(红包消息，卡片消息等等的发送依赖于iOS的tweak)

## UUID
 相信大家都已经很清楚微信的登录过程了 这里我再啰嗦一下

首先我们会去微信服务器请求一个`uuid`, 拿到这个`uuid`我们需要让微信app确认这个uuid完成登录过程。

我们的机器人很可能是运行在自己的pc上或者服务器上，那么掉线以后我们怎么知道呢？ 怎么及时让他上线呢？

我们需要及时的用手机来扫描这个uuid， 比如通过邮件发出来？ 通过短信发出来？ 或者通过iOS tweak自动扫码。

这里我提供了一个接口 `UUIDProcessor` 大家可以实现这个接口，去自定义自己的uuid处理方式

目前包中支持3种处理

    1. 使用本机的图片浏览工具打开二维码，然后掏出手机扫描
    2. 直接将二维码打印在命令行中，然后用掏出手机扫描
    3. 将uuid发送到tweak然后微信自动识别

你可以实现譬如这样的：

    _4. 将二维码发送到自己的QQ邮箱，利用QQ的邮箱提醒功能来及时登录_
    _5. 你可以将uuid通过短信发送到手机，然后再想办法通过微信扫描_

## 模块

好了这下到了最重要的部分了，`wechat`包本身并不做任何的业务逻辑。他做的仅仅是 **维持和服务器的连接，然后将服务器的事件分发出来**。所有的业务都由我们自定义的`module`来处理。所以说我们基本都是在开发自定义`module`的路上或者已经死在了路上~~

这里我们拉代码出来看看 我们支持分发那些数据

首先是登录态的变化 `LoginStateModule` 接口一共定义了2个方法
``` go
	WechatDidLogin(wechat *WeChat)
	WechatDidLogout(wechat *WeChat)
```
这个2个方法在机器人第一登录和掉线后自动登录时都会调用，我们自己的`module`实现这个接口以后可以在`WechatDidLogin`中拿到wechat对象，就可以愉快的做很多坏事啦

接着是处理新消息的模块 `MsgModule` 也只有2个方法简单快捷

```go
	MapMsgs(msg *CountedContent)
	HandleMsgs(msg *CountedContent)
```

这里的`CountedContent`也是包定义的一个结构体，在这里我们不妨也看一下他的定义

```go
type CountedContent struct {
	Count   int
	Content []map[string]interface{}
}
```
就像他的名字，他有一个`Count`有一个`Content`其中Content就是从服务器返回的json数据。

我们接着回头来看这个接口`MapMsgs`用来对所有的消息进行一个预处理，举个例子
我们收到的红包提醒是这样的
```json
{type = 1000, content = "Receive a redpacket,view on phone" }
```
由于网页端不支持红包消息，所以是以一个系统通知发过来的，我们可以把这条消息`map`为
```json
{type = 99, content= "大红包，卧槽大红包 好激动" }
```
在`HandleMsgs`方法中我们就可以判断只要`type=99` 就知道我们收到红包啦

来看最后一个接口`ContactModule`
```go
	MapContact(contact map[string]interface{})
	HandleContact(contact map[string]interface{})
```
和`MsgModule`极其的相似，我们就不细说了。

## 现在已有的模块

现在我们已经有了一些功能模块，你可以在这里找到，并且每个模块都有自己的安装说明

### bridge
	这是和iOS Tweak配合使用的模块，现在大家应该用不到，等iOS Tweak开源以后这里会补充起来 //TODO

### convenience
这是一个为处理消息提供便利的包，使得开发这不需要自己去定义`Module`就可以处理特定的消息

##### step 1 安装包
```go
go get github.com/KevinGong2013/ggbot/modules/convenience
```
##### step 2 导入包
```go
import "github.com/KevinGong2013/ggbot/modules/convenience"
```
##### step 3 注册module
```go
wxbot.RegisterModule(convenience.DefaultMsgStream)
```
#### step 4 处理消息
```go
// 一共支持一下4种注册方式

// /msg/solo／GGBot 单聊的GGBot的消息
// /msg/solo/ 剩下的所有单聊消息
// /msg/group/GGBot测试群 来自GGBot测试群的消息
// /msg/group 剩下的所有群聊消息

// 比如我想对我老婆发给我的每一条消息都回复一句 `小的知道了`
convenience.Handle(`/msg/solo/老婆`, func(msg map[string]interface{}) {
		un, _ := wxbot.UserNameByNickName(`老婆`)
		wxbot.SendTextMsg(`小的知道了`, un)
	})
// 其他人的消息都打印到日志
convenience.Handle(`/msg/solo`, func(msg map[string]interface{}) {
		logger.Debugf(`%v`, msg)
	})
// 如果是群聊中收到老婆的消息，就回一句 `老子收到了`
convenience.Handle(`/msg/group/老婆`, func(msg map[string]interface{}) {
		un, _ := wxbot.UserNameByNickName(`老婆`)
		wxbot.SendTextMsg(`老子收到了`, un)
	})
// 群聊中`机器人`的消息就会一句`好巧呀, 我也是机器人`
convenience.Handle(`/msg/group/机器人`, func(msg map[string]interface{}) {
		un, _ := wxbot.UserNameByNickName(`机器人`)
		wxbot.SendTextMsg(`好巧呀, 我也是机器人`, un)
	})
```
### echo
这个模块是将所有的服务器时间打印到日志系统，安装导入和上一个包一样，我们来看注册

```go
wxbot.RegisterModule(new(echo.Echo)) // 很简单吧
```
### gguuid
这是用来把二维码直接绘制在`terminal`的模块，安装导入和上一个包一样，我们来看注册

```go
wxbot, err := wechat.WeekUp(&gguuid.UUIDProcessor{})
```
### media
这个模块是在收到多媒体消息以后，自动下载媒体文件，并且以消息ID为文件名放方便我们索引， 安装导入和上一个包一样，我们来看注册

```go
// mediaPath 媒体文件的下载路径。 推荐 `.ggbot/media`
d, err := media.NewDownloader(mediaPath)
if err == nil {
	wxbot.RegisterModule(d)
} else {
	logger.Error(err)
}
```
### notice
这是一个`demo`模块，创建一个webhook，将git 服务器的时间转发到微信群或者联系人，现在只有基础功能 正在开发中，安装导入和上一个包一样，我们来看注册

```go
// 监听到git服务器消息以后发送到 `filehelper`
wxbot.RegisterModule(notice.NewNotice(`filehelper`))

// 需要创建一个webhook所以需要listen一个端口， 对应的需要在git服务器的
// webhook连接中填写对应的地址
logger.Error(http.ListenAndServe(`:3280`, nil))
```

### redpacket
红包处理的模块，可以后台自动抢红吧，分析红包数据 这个模块需要依赖`bridge`模块 iOS Tweak 开源以后这里会补充起来 // TODO

### storage
本地存储模块，用于将联系人信息和消息以`json`文件的方式保存起来，安装导入和上一个包一样，我们来看注册

```go
// json 文件存放的根目录。推荐`.ggbot/db`
st, err := storage.NewStorage(dbPath)
if err == nil {
	wxbot.RegisterModule(st)
} else {
	logger.Error(err)
}
```
### tuling
图灵机器人模块，用于接入图灵机器人api，目前支持回复群聊和单聊的文本消息，没有判断@，还比较傻 欢迎大家提pr，安装导入和上一个包一样，我们来看注册

```go
// apikey 需要大家去图灵机器人官网注册 http://www.tuling123.com
wxbot.RegisterModule(tuling.NewBrain(apikey))
```
### ui
这个模块比较大，提供一个基于`terminal`的用户界面，目前支持一些基本信息的展现，还不完善 欢迎大家提pr。注意 `ui`模块和`gguuid`有冲突不能同时使用，安装导入和上一个包一样，我们来看注册

```go
// 这里的mediaPath用来监听本地媒体文件的变化，推荐配合 media 模块使用
ui := ui.NewUI(mediaPath)

wxbot.RegisterModule(ui)
ui.Loop() // 注意ui.Loop() 会阻塞调用
```

## TODO

- []完善现有模块
- []添加更多模块
- []添加单元测试

## 交流讨论

	1.github issue (推荐)
	2.qq 群：609776708

## License

    The code in this repository is licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

**NOTE**: This software depends on other packages that may be licensed under different open source licenses.

Copyright 2017 - 2027 Kevin.Gong  aoxianglele#icloud.com
