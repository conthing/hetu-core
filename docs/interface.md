# 上报下发接口

<a name="tSzvi"></a>
### 概述
系统提供以下上报下发接口，

1. HTTP下发
1. HTTP上报（可设置HTTP Address、端口、URL、打开关闭）
1. MQTT下发（可设置MQTT broker Address、端口、打开关闭）
1. MQTT上报（设置和3一样）

其中1、3的命令格式相同，2、4上报内容格式相同。<br />**MQTT 上报**<br />MQTT 上报主题为 `/hetu/${SN}/report`<br />**MQTT 下发**<br />MQTT 下发主题为 `/hetu/${SN}/command`<br />**HTTP 上报**<br />HTTP 上报的 URL 由用户在页面设置<br />**HTTP 下发**<br />HTTP 下发的 URL 为 http://192.168.0.101/hetu/command<br />**内容格式**<br />内容格式示例如下。
```json
{
  "type": "xxx",
	"data":{
  	...
  }
}
```
type 为数据格式的识别符，定义为“zigbee”或二次开发应用的识别符。当 type 为“zigbee” 时，用于表示 zigbee 报文的上传与下发。二次开发应用不能使用“zigbee”作为识别符

<a name="HiUkH"></a>
### Redis
二次开发应用与远程服务器端的通信均通过 Redis Pub/Sub 功能实现。

收到下行命令后，如果命令内容的 type 不是 “zigbee” 则发布到 Redis 队列。

下行的通信 Redis topic 为 下行内容中的type，即二次开发应用的识别符。

hetu-core 程序会订阅 “hetu-core" 主题的 Redis 队列，将内容通过 MQTT 或 HTTP 上报。所以二次开发的应用上行的通信需要发布到 “hetu-core” Redis topic。

具体示例如下。
<a name="Y2RoS"></a>
### MQTT 上报 
<a name="PEu7L"></a>
#### Zigbee 设备 -> hetu-core -> MQTT 服务器 -> 远程服务器

- MQTT 主题为 `/hetu/${MAC}/report`

- 报文示例如下

```json
{
  "type":"zigbee",
  "data":
  	{
    	"eui64":5149013401532256,
    	"addr":15,
    	"message":"DwAgsAAPAACcIwAAAH8AAAAAAAAAAG1vAG8AACEhAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAmWg==",
      "time":"2016-02-12T02:29:32.450819869+08:00",
  	}
}
```

<a name="bVFpQ"></a>
#### 二次开发应用 -> Redis 队列 -> hetu-core -> MQTT 服务器 -> 远程服务器

- Redis 队列主题为 “**hetu-core”**
- **type **不能是 “zigbee” ，为二次开发应用的识别符
- MQTT 主题为 `/hetu/${MAC}/report`
- Redis 队列报文示例如下图

```json
{
  "type": "your_app_name",
	"data":{
  	...
  }
}
```

<a name="ezojS"></a>
### HTTP 上报
<br />
<a name="U0bnE"></a>
#### Zigbee 设备 -> hetu-core -> 远程服务器（HTTP服务）

- 报文示例如下

```json
{
  "type":"zigbee",
  "data":
    {
        "eui64":5149013401532256,
        "addr":15,
        "message":"DwAgsAAPAACcIwAAAH8AAAAAAAAAAG1vAG8AACEhAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAmWg==",
     	  "time":"2016-02-12T02:29:32.450819869+08:00",
    }
}
```

<a name="QFl1E"></a>
#### 二次开发应用 -> Redis 队列 -> hetu-core -> 远程服务器（HTTP服务）

- Redis 队列主题为“hetu-core”
- **type **不能是 “zigbee” ，为二次开发应用的识别符
- 通过 POST 方法上报给 HTTP 服务器
- **Redis 队列报文示例**如下图

```json
{
  "type":"xxx",
	"data":{}
}
```

<a name="393Uj"></a>
### MQTT 下发
<a name="8n8fL"></a>
#### 远程服务器 -> MQTT 服务器 -> hetu-core -> Zigbee 设备

- MQTT 主题为 `/hetu/${MAC}/command`
- 报文中的 type 为 zigbee
- hetu-core会把报文中的 data 转发给 zigbee 设备
- 报文示例如下

```json
{
	"type": "zigbee",
  "data": {}
}
```

<a name="Ft77R"></a>
#### 远程服务器 -> MQTT 服务器 -> hetu-core -> Redis 队列 -> 二次开发应用

- MQTT 主题为 `/hetu/${MAC}/command`
- **type **不能是 “zigbee” ，为二次开发应用的识别符
- 将报文中的 type 作为 redis 队列的主题
- hetu-core会根据报文中的 type 把 data 转发给 redis 队列
- 报文示例如下

```json
{
	"type": "your_app_name",
  "data": {}
}
```

<a name="ylRO1"></a>
### HTTP 下发
<a name="nI9oV"></a>
#### 远程服务器 -> hetu-core -> Zigbee 设备

- API 示例 : `http://192.168.0.101/hetu/**command**`
- 报文中的 type 为 zigbee
- hetu-core会把报文中的 data 转发给 zigbee 设备
- 报文示例如下

```json
{
	"type": "zigbee",
  "data": {}
}
```

<a name="kqedk"></a>
#### 远程服务器 -> hetu-core -> Redis 队列 -> 二次开发应用

- API 示例 : `http://192.168.0.101/hetu/**command**`
- hetu-core会根据报文中的 topic 把报文中的 data 转发给 Redis 队列
- 报文示例如下

```json
{
	"type": "your_app_name",
  "data": {}
}
```
