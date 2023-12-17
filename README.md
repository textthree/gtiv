## 简介
GTIV 是一个短视频社交项目。包含短视频、直播、即时通讯、音视频通话、单聊、群聊等功能。

此项目源于作者在技术学习探索过程中写的一些试验代码，在空闲之际将其整理开源，并写了一些简单的业务 Demo 形成一个可运行的产品。
包含后 [GTIV后端](https://github.com/text3cn/gtiv) 加 [GTIV Android](https://github.com/text3cn/gtiv-android) 端，有这方面技术需求的朋友可以将其作为一个参考。

由于作者精力有限，一些 Demo 怎么快怎么写，并没有严格的编码规约，如果你有自己的编码风格，请自行改造。
同时也希望您踊跃参与开源贡献或测试 bug 提出 Issue，愿您在技术的道路上突飞猛进，祝您工作顺利！

## 技术栈
采用微服务构架，由于个人开发为了方便，将多个服务放在一个项目大仓中。

- 使用 [Goodle](github.com/text3cn/goodle) 开发框架，这是我的另一个开源项目。
- [Goim](https://github.com/Terry-Mao/goim) 即时消息推送
- [Livego](https://github.com/gwuhaolin/livego) 直播推拉流
- [Turn](https://github.com/pion/turn) WebRTC 音视频通话中继服务
- 使用 [Gorm](https://github.com/go-gorm/gorm) 操作数据库存
- 使用 [Swaggo](https://github.com/swaggo/swag) 生成 [Opemapi](http://goodle.text3.cn/docs/openapi/) 文档
- 使用 [FFmpeg](https://github.com/FFmpeg/FFmpeg) 处理视频
- 使用七牛云存储静态资源
- 使用 Kafka 作为消息队列
- 使用 Redis 做缓存


由于 Goim 开源出来就不怎么维护了，GTIV 对 Goim 进行了改造整合:
1. 将 Discovery 改为使用 Etcd 进行服务发现，以更好的适应 K8S 部署。
2. 将 Http 服务的部分由 Gin 框架改为使用 Goodle 框架开发。
3. 将 Redigo 改为 Goodle 框架提供的 [Redis 开发套件](http://goodle.text3.cn/docs/kit/redis)。

## 服务架构
### 有状态服务
这些有状态服务通常没有太多需要迭代发版的逻辑。

#### comet
长连接接入网关，可以部署大量低廉配置机器的边缘节点就近接入。每个节点需要一个唯一的 server.id，
目前是通过 Golang `os.Hostname()`，注意一定要保证不同的 comet 进程值唯一。

#### job
消息队列的消费者，IM 消息通过 job 服务消费 kafka 队列进行推送。
job 可以根据 kafka 的 partition 来扩展多 job 工作方式，具体可以参考 kafka 的 partition 负载。

#### livego
直播推拉流服务，livego 目前的性能是不错的，得益于 Golang 的天然并发，推拉流占用资源不大，生产环境拉流推荐使用 CDN 满足大流量需求。
这有一篇 [推拉流方案总结](http://text3.cn/blog-343333788.html) 的文章可供参考。

#### webretc
音视频通话的信令服务及中继服务。在美国的网络环境 webrtc 通过信令服务器能打通 80% 左右的通话连接。
国内的网络环境就比较难于连通，三大运营商以及各种教育网、内网、防火墙等。可能 40% 的连通率都不到，需要中继服务器提供流量转发。

#### etcd
生产环境建议使用 k8s 部署你的项目，自带 etcd 集群。

### 无状态服务
日常需要频繁迭代的需求通常都在这两个服务里边进行开发。

#### imbiz 
IM 相关逻辑层的 Http 服务。

#### videobiz 
短视频相关业务逻辑的 Http 服务。


### 其他
- kafka 可以使用多 broker 或者多 partition 来扩展队列
- router 属于有状态节点，imbiz 可以使用一致性 has h配置节点，增加多个 router 节点（目前还不支持动态扩容），提前预估好在线和压力情况


## 开始开发
### 作者本地环境
- macOS 12.3.1
- go 1.20.6
- redis 7.2.3
- etcd 3.4.27
- kafka 2.13-2.8.1
- mysql 8.2.0

Mysql 设置允许 ONLY_FULL_GROUP_BY
 
### 安装
```bash
go get github.com/text3cn/gtiv
```
### 编译
执行 Makefile 指令：
```bash
make build
```
执行命令后会在项目根目录生成一个 `dist` 目录，所有服务都会编译到这里。
注意服务之间使用 Etcd 进行服务发现，在启动服务前需要确保你本地安装了 Etcd。 
在 `/scripts/etcd/` 中附有 Etcd 的配置文件和 Linux、Mac 环境下的安装脚本。
>注意：<br>
> 1、Makefile 中调用的是 /scripts/build.sh 脚本，请确保你本机的 build.sh 具有可执行权限。<br/><br/>
> 2、由于 Etcd 通常用于 Linux 环境，Windows 上的 etcd 安装相对较少见，而且 Etcd 官方对 Windows 的支持有限，Windows 用户建议搭建 Docker 环境来运行 GTIV 项目。

### 数据库
Mysql 表结构通过 Gorm 自动迁移，配置好数据库，运行相关服务会自动创建表。

### 服务端端口清单
2002 imbiz rpc <br/>
2003 comet rpc <br/>
2004 comet tcp <br/>
2005 comet websocket <br/>
2006 comet websocket tlsBind <br/>
2007 pionturn <br/>
2009 videobiz http <br/>
2010 livego http <br/>
2011 livego rtmp <br/>
2012 livego hls <br/>
2013 livego flv <br/>
2014 singnalServer <br/>
2015 imbiz http <br/>

 
## 声明与支持
此项目属于个人开源作品，仅做学习交流使用，因为个人精力有限，并未做大量测试。
如果您需要将其作为商业用途，需考虑此项目可能存在的 Bug，以及相关第三方开源库的开源许可证条款。

如果你需要此项目相关的技术咨询，可以给我来杯咖啡。微信号: text3cn
