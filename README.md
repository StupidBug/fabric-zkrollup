# ZK Rollup Sidechain

一个专注于基础区块链功能和ERC20类操作的ZK Rollup侧链实现。

## 项目结构

```
zkrollup/
├── cmd/                 # 命令行入口
│   ├── zkrollup/       # 主程序入口
│   └── keygen/         # 密钥生成和交易签名工具
├── internal/           # 内部包
│   ├── api/           # HTTP API处理器
│   │   ├── handlers/  # API请求处理器
│   │   ├── middleware/# 中间件
│   │   └── router/    # 路由配置
│   ├── core/          # 核心功能
│   │   ├── blockchain/# 区块链核心实现
│   │   ├── txpool/    # 交易池管理
│   │   └── crypto/    # 加密相关功能
│   ├── types/         # 数据类型定义
│   │   ├── block/     # 区块相关类型
│   │   ├── transaction/# 交易相关类型
│   │   └── state/     # 状态相关类型
│   └── zk/            # ZK证明相关功能
├── pkg/               # 可导出的包
│   └── utils/        # 通用工具函数
├── scripts/          # 部署和管理脚本
│   ├── deploy.sh     # 部署脚本
│   └── test.sh       # 测试脚本
├── test/             # 测试文件和脚本
│   ├── integration/  # 集成测试
│   └── transfer_test.sh # 转账测试脚本
├── docs/             # 项目文档
│   ├── api.md        # API文档
│   └── design.md     # 设计文档
└── .gitignore       # Git忽略文件配置
```

## 主要功能

### 已实现功能
- 基础区块链结构
  - Merkle树实现（支持空交易块）
  - 哈希计算
  - 区块链状态管理
- 账户管理系统
  - 账户余额管理
  - 交易池管理
  - 状态更新
- ERC20类代币操作
  - 转账交易
  - 余额查询
- 区块管理
  - 区块创建和验证（包含空区块处理）
  - 交易打包
  - Merkle根计算
  - 状态更新
- RESTful API接口（基于Gin框架）
  - 交易处理
  - 余额查询
  - 区块操作
- 完整的交易跟踪和查询系统
- 密钥管理工具
  - 密钥对生成
  - 交易签名
  - 公钥验证
- 交易签名系统
  - ECDSA签名生成和验证
  - 交易安全性验证
  - 公钥管理

### 技术特性
- 支持空交易区块的创建和验证
- 自动区块创建
  - 可配置的区块生成间隔（默认5秒）
  - 智能区块打包（仅在有交易时创建区块）
  - 可动态开启/关闭自动打包功能
- Merkle树优化
  - 自动处理空交易列表
  - 奇数个交易节点的处理（复制最后一个节点）
  - 空区块使用空哈希作为Merkle根
- 线程安全的状态管理
  - 使用互斥锁保护共享资源
  - 支持并发交易处理
- ECDSA签名系统
  - 基于P256曲线的密钥生成
  - 交易签名和验证
  - 公钥管理

### 待实现功能
- 持久化存储
- ZK证明系统
- 高级查询功能
- 性能优化
- 安全增强
- 监控和日志系统

## 开始使用

### 环境要求

- Go 1.20 或更高版本
- jq (用于运行测试脚本)
- curl (用于API测试)

### 安装

```bash
git clone https://github.com/yourusername/zkrollup.git
cd zkrollup
go mod download
```

### 构建和运行

你可以选择手动构建运行或使用部署脚本：

#### 手动构建运行

```bash
# 构建主程序
go build ./cmd/zkrollup

# 构建密钥生成工具
go build ./cmd/keygen

# 运行主程序
./zkrollup
```

#### 使用密钥生成工具

```bash
# 生成新的密钥对
./keygen -genkey

# 签名交易
./keygen -sign -from <sender_address> -to <receiver_address> -value <amount> -nonce <nonce> -privkey <private_key>
```

#### 使用部署脚本

部署脚本提供了自动化的部署流程：

```bash
# 设置脚本执行权限（首次使用时）
chmod +x scripts/deploy.sh test/transfer_test.sh

# 使用默认端口(8080)
./scripts/deploy.sh

# 指定自定义端口
./scripts/deploy.sh -p 8081
```

部署脚本功能：
- 自动停止已运行的实例（如果有）
- 编译项目
- 启动服务器
- 验证服务器是否正常运行
- 支持自定义端口配置

### 运行测试

测试脚本提供了简洁高效的功能测试流程。所有脚本都应在项目根目录下运行：

```bash
# 运行测试脚本
./test/transfer_test.sh
```

测试脚本功能：
1. 自动检查并启动服务器（如果未运行）
2. 执行以下核心测试流程：
   - 设置初始账户余额
   - 验证账户余额
   - 执行转账交易
   - 等待交易确认（最多60秒）
   - 验证转账后的余额变化
   - 查询所有区块和交易信息

测试输出使用彩色标记：
- 绿色：测试步骤和成功信息
- 红色：错误信息
- 默认：测试结果和详细数据

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