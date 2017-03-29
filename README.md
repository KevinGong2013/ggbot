# GGBot

一个牛逼的微信机器人，非常适合非技术人员自由定制
### 先看个视频
[GGBot 演示](http://7xnowv.com1.z0.glb.clouddn.com/ggbot-demo.mp4)

### 下载地址

[window 版本](http://oap91rhcb.bkt.clouddn.com/GGBot-windows.exe)

> mac os x 和 linux 建议自己编译

### 自定义

直接运行下载到的可执行文件即可，弹出二维码以后用手机扫描二维码即可登陆。 机器人自动加载：`微软小冰`  `加群欢迎语`  `群签到`  `添加好友自动通过` 。更多的功能正在开发中，欢迎在`issue`中提出你想要的功能。

> 默认会加载`微软小冰`， 所以请大家先用`微信app`关注微软小冰公众号

##### 如何使用图灵机器人
机器人在运行过一次以后会在可执行文件的统一目录生成 `conf.yaml`,  打开这个文件
``` yaml
showqrcodeonterminal: false #是否在命令行中显示二维码
fuzzydiff: true #联系人对比是否启用模糊匹配
uniquegroupmember: true #是否为群成员生成ggid
features:
  assistant:
    enable: false
    ownerggid: ""
  guard:
    enable: false
  tuling:
    enable: false
    key: ""
  xiaoice:
    enable: false
  webhookservice:
    enable: false
    msgwebhook: ""
    loginstatewebhook: ""
    uuidwebhook: ""
```
修改下面2行
``` yaml
  tuling:
    apikey: "" #这里填写你申请到的图灵机器人key
    enable: true # 将这里改为true
```

> apikey 可以在图灵机器人的官网免费申请 [点击这里立即申请](http://www.tuling123.com)

### 二次开发
本机器人基于[wechat](https://github.com/KevinGong2013/wechat)开发
###### 安装`go-lang`开发环境
[传送门](https://www.golang.org)

###### 安装`wechat`包
``` go
go get github.com/KevinGong2013/wechat
```
###### clone 源码
``` go
git clone https://github.com/KevinGong2013/ggbot.git
```
###### 编译运行
``` bash
cd ggbot
go build
./ggbot
```

### 交流讨论

	1.github issue (推荐)
	2.qq 群：609776708

### 常见问题

##### 0x00 
  Q: windows 系统编译运行，cmd显示不正常改怎么办？  
  A: [Enable ANSI colors in Windows command prompt](https://web.liferay.com/web/igor.spasic/blog/-/blogs/enable-ansi-colors-in-windows-command-prompt)

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
