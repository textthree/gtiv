
<img src='./logo.png' width='200px' height='80px'/>

docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -p 8090:8090 -d gwuhaolin/livego

[![Test](https://github.com/gwuhaolin/livego/workflows/Test/badge.svg)](https://github.com/gwuhaolin/livego/actions?query=workflow%3ATest)
 
## 简介
### 支持的传输协议
- RTMP
- AMF
- HLS
- HTTP-FLV

### 支持的容器格式
- FLV
- TS

### 支持的编码格式
- H264
- AAC
- MP3

## 安装
### 从源码编译
1. 下载源码`git clone https://github.com/text3cn/livego`
2. 编译`go build`

### 启动 
```bash
./livego
```
配置项在`livego.yaml`进行配置修改，或启动时指定配置
```bash
./livego  -h
Usage of ./livego:
      --api_addr string       HTTP管理访问监听地址 (default ":2010")
      --config_file string    配置文件路径 (默认 "livego.yaml")
      --flv_dir string        输出的 flv 文件路径 flvDir/APP/KEY_TIME.flv (默认 "tmp")
      --gop_num int           gop 数量 (default 1)
      --hls_addr string       HLS 服务监听地址 (默认 ":2012")
      --hls_keep_after_end    Maintains the HLS after the stream ends
      --httpflv_addr string   HTTP-FLV server listen address (默认 ":2013")
      --level string          日志等级 (默认 "info")
      --read_timeout int      读超时时间 (默认 10)
      --rtmp_addr string      RTMP 服务监听地址 (默认 ":2011")
      --write_timeout int     写超时时间 (默认 10)
```

## 使用
以两路推流为例。

### 推流
1. 访问 http 接口获取推流 channelkey
```bash
# room 参数为自定义房间号
http://localhost:2010/control/get?room=room1
http://localhost:2010/control/get?room=room2
```
2. 推流，例如使用 ffmpeg 推流
```bash
# appname 默认是 live，-stream_loop -1 代表无限循环推流
ffmpeg -re -stream_loop -1 -i demo.flv -c copy -f flv rtmp://localhost:2011/{appname}/{channelkey}
```
源码中 videos 目录有提供两个测试用视频文件。

### 拉流
支持多种播放协议，播放地址如下:
- `RTMP`:`rtmp://localhost:2011/{appname}/room1`
- `HLS`:`http://127.0.0.1:2012/{appname}/room1.m3u8`
- `FLV`:`http://127.0.0.1:2013/{appname}/room1.flv`

- `RTMP`:`rtmp://localhost:2011/{appname}/room2`
- `HLS`:`http://127.0.0.1:2012/{appname}/room2.m3u8`
- `FLV`:`http://127.0.0.1:2013/{appname}/room2.flv`


### 从 Docker 启动
```
docker run -p 2011:1935 -p 2013:7001 -p 2012:7002 -p 2010:8090 -d gwuhaolin/livego
```


### [和 flv.js 搭配使用](https://github.com/gwuhaolin/blog/issues/3)


