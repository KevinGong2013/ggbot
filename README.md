#GGBot

一个牛逼的微信机器人，非常适合非技术人员自由定制

### 下载地址

[window 版本](https://github.com/KevinGong2013/ggbot/raw/master/release/GGBot-windows.exe)

> mac os x 和 linux 建议自己编译

### 自定义

直接运行下载到的可执行文件即可，弹出二维码以后用手机扫描二维码即可登陆。 机器人自动加载：`微软小冰`  `加群欢迎语`  `群签到`  `添加好友自动通过` 。更多的功能正在开发中，欢迎在`issue`中提出你想要的功能。

> 默认会加载`微软小冰`， 所以请大家先用手机app关注微软小冰公众号

##### 如何使用图灵机器人
机器人在运行过一次以后会在可执行文件的统一目录生成 `conf.yaml`,  打开这个文件
``` yaml
features:
  assistant:
    enable: true
  guard:
    enable: true
  tuling:
    apikey: "" #这里填写你申请到的图灵机器人key
    enable: true # 将这里改为true
  xiaoice:
    enable: true
```
修改下面2行
``` yaml
  tuling:
    apikey: "" #这里填写你申请到的图灵机器人key
    enable: true # 将这里改为true
```

> apikey 可以在图灵机器人的官网免费申请 [点击这里立即申请](http://www.tuling123.com)

###二次开发
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
