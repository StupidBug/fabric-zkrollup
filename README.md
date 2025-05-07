# Fabric ZK Rollup Sidechain

一个专注于基础区块链功能和ERC20类操作的ZK Rollup侧链实现，基于Hyperledger Fabric。

## 开始使用

### 安装

```bash
git clone https://github.com/StupidBug/fabric-zkrollup.git
cd fabric-zkrollup
go mod download
```

### 构建和运行

部署脚本提供了自动化的部署流程：

```bash
# 部署存证zk组件
make run TARGET=storage


# 部署资产zk组件
make run TARGET=transfer
```

## API文档

### 交易相关接口

#### 发送交易
- **POST** `/api/v1/transaction/send`
- **请求体**:
  ```json
  {
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": 100,
    "nonce": 1,
    "signature": {
      "r": "hex_string",
      "s": "hex_string"
    },
    "publicKey": {
      "x": "hex_string",
      "y": "hex_string"
    }
  }
  ```
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "hash": "hex_string",
      "from": "address",
      "to": "address",
      "value": 100,
      "nonce": 1,
      "status": "pending",
      "timestamp": 1234567890,
      "signature": {
        "r": "hex_string",
        "s": "hex_string"
      },
      "publicKey": {
        "x": "hex_string",
        "y": "hex_string"
      }
    }
  }
  ```

#### 查询交易
- **GET** `/api/v1/transaction/get?hash={transaction_hash}`
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "hash": "hex_string",
      "from": "address",
      "to": "address",
      "value": 100,
      "nonce": 1,
      "status": "confirmed",
      "timestamp": 1234567890
    }
  }
  ```

### 账户相关接口

#### 查询余额
- **GET** `/api/v1/balance/get?address={address}`
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "address": "address",
      "balance": 1000
    }
  }
  ```

#### 查询Nonce
- **GET** `/api/v1/account/nonce?address={address}`
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "address": "address",
      "nonce": 1
    }
  }
  ```

### 状态相关接口

#### 查询状态根
- **GET** `/api/v1/state/root`
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "stateRoot": "hex_string"
    }
  }
  ```

### 区块相关接口

#### 查询所有区块
- **GET** `/api/v1/blocks`
- **响应**:
  ```json
  {
    "status": "success",
    "data": {
      "blocks": [
        {
          "height": 0,
          "hash": "hex_string",
          "prevHash": "hex_string",
          "merkleRoot": "hex_string",
          "stateRoot": "hex_string",
          "timestamp": 1234567890,
          "transactionCount": 0,
          "transactions": []
        },
        {
          "height": 1,
          "hash": "hex_string",
          "prevHash": "hex_string",
          "merkleRoot": "hex_string",
          "stateRoot": "hex_string",
          "timestamp": 1234567891,
          "transactionCount": 1,
          "transactions": [
            {
              "hash": "hex_string",
              "from": "address",
              "to": "address",
              "value": 100,
              "nonce": 0,
              "status": "confirmed",
              "timestamp": 1234567890
            }
          ]
        }
      ]
    }
  }
  ```

详细的API文档请参考 [API文档](docs/api.md)