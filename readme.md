2018/11/17 log
1、交易添加ID

2、增加定时自动出块

3、修改交易池，以适应交易检查和进出交易池的要求

4、创世区块指定两个账号的余额，一开始转账的时候必须用这两个账号(1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY, 1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq)

5、启动时需要制定账号，并且使用独立的钱包，因为需要钱包检查密钥


2018/10/28 log

1、添加transcation 和 state，支持transcation驱动下的state变化

2、支持transcation的p2p传播，支持state的同步

2018/10/23 log

1、支持POS共识算法，节点间同步节点的抵押信息和新块，最后由pickwinner函数选出winner

2、pickwinner实现了POS共识算法，支持分布式伪随机确定数值的生成，随机源来着上一个区块的时间戳

3、支持多个节点达成共识，链启动时需要带上账号信息(-a)，作为出块者标识

4、节点增多，pickwinner休眠的时间需要相应的增多，因为手输result需要时间，尽量在一个出块周期内输完所有result

## 操作说明


### 编译可执行文件
项目cmd 目录执行go build


### 使用可执行文件生成新账户
./cmd -c account lzhx_ createwallet     (lzhx_ 表示钱包后缀)

Your new address: 1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY

./cmd -c account zhong_ createwallet     (lzhx_ 表示钱包后缀)

Your new address: 1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq

### 使用可执行文件查看钱包地址列表
./cmd -c account lzhx_ listaddresses  

13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De 

1LcubHwTs7AuHGxoPdwttbcbJkrrCwWUYX 

1zmkxjXmisf4kvQJCUnhKRYMbpsXBdPYf 

1EaKYX3U3GGjM4NoNXNVhMk7KQ1UKnTXgz 

1Pz7MTSoESDmnQDMaeLijG4qCbsPdvrus6 

### 启动链并连

接对端节点
./cmd -c chain -s lzhx_ -l 8080 -a 1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY

-s 表示 钱包后缀

-a 表示钱包地址


启动当前节点后日志输出一下信息

local http server listening on 127.0.0.1:8081
2018/09/21 15:06:20 I am /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj
2018/09/21 15:06:20 Now run "go run main.go -c chain -l 8082 -d /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj" on a different terminal

在另一个terminal中启动对端节点（加入-s参数表示该节点的钱包后缀名）

./cmd -s lzhx_ -c chain -l 8082 -d /ip4/127.0.0.1/tcp/8080/ipfs/QmdhJPDZaLPCFjZMsuLfVtzZMNaZMPp6wT85gYdRnVcppj -a 1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq -s zhong_

两节点正常连接后可在terminal中输入任意数字，该操作将产生新的块并同步块信息到对端节点

### 查看链状态、发送交易及交易打包

1.通过get形式的http请求查看链信息
e.g

path: http://127.0.0.1: &lt; port &gt;

return:
```json
   [
     {
       "index": 0,
       "timestamp": 1540610566,
       "result": 0,
       "validator": "1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY",
       "hash": "2ac9a6746aca543af8dff39894cfe8173afba21eb01c6fae33d52947222855ef",
       "prevhash": "",
       "proof": 0,
       "transactions": null,
       "CoinBase": {
         "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq": {
           "Addr": "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq",
           "Nonce": 0,
           "Balance": 10000
         },
         "1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY": {
           "Addr": "1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY",
           "Nonce": 0,
           "Balance": 10000
         }
       }
     },
     {
       "index": 1,
       "timestamp": 1540737336,
       "result": 2,
       "validator": "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq",
       "hash": "25db28fe02a65b704d7770eebe6b9b8f6ac162b0263620384a3d5b75974f0cb1",
       "prevhash": "2ac9a6746aca543af8dff39894cfe8173afba21eb01c6fae33d52947222855ef",
       "proof": 0,
       "transactions": null,
       "CoinBase": null
     }
   ]
```


2.通过post形式的http接口发送交易到链上
e.g

path:   http://127.0.0.1: &lt; port &gt; /txpool

param:

```json
    {
        "From": "1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY",
        "To": "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq",
        "Value": 100,
        "nonce": 3,
        "Data": "message"
    }
```

return:
```json
    {
      "sender": "1KSKahQT9n69sgqn4aVmRUPpydf6AUeeZY",
      "recipient": "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq",
      "amount": 100,
      "account_nonce": 3,
      "data": "bWVzc2FnZQ=="
    }
```



3.通过post形式的http接口发送信息产生新块
e.g

path:   http://127.0.0.1: &lt; port &gt; /block

param:

```json
    {"Msg": 123}
```

return:
```json
    {
      "index": 2,
      "timestamp": "2018-09-20 18:03:23.460148402 +0800 CST m=+24.501698347",
      "result": 123,
      "hash": "0ee7933883ae99f99fdc964042008426240066408ef8f0598e780a8158202f68",
      "prevhash": "e792220c169142a4561b7320005716a636a27b25bb2cb03c409a20ef64037d53",
      "proof": 0,
      "transactions": [
        {
          "amount": 1,
          "recipient": "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm",
          "sender": "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De8",
          "data": "bWVzc2FnZQ=="
        }
      ],
      "accounts": {
        "13qAPhDtk82VdLMcaUoh7jwNi5HpFX6De": 9999,
        "17eeNAJcUWECkHLDgGcXwZPKrYteNLq2hm": 1
      }
    }
```


4.通过post形式的http接口查看账户余额
e.g

path:   http://127.0.0.1: &lt; port &gt; /getbalance

param:

```json
    {"Address": "1EFnWYm1suorEdt5XLEJ9UMTYQjGzqmiJq"}
```

return:
```
    1
```
