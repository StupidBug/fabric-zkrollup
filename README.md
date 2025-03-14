# ZK Rollup Sidechain

一个专注于基础区块链功能和ERC20类操作的ZK Rollup侧链实现。

## 项目结构

```
zkrollup/
├── cmd/                 # 命令行入口
│   └── zkrollup/       # 主程序入口
├── internal/           # 内部包
│   ├── api/           # HTTP API处理器
│   │   ├── handlers/  # API请求处理器
│   │   ├── middleware/# 中间件
│   │   └── router/    # 路由配置
│   ├── core/          # 核心功能
│   │   ├── blockchain/# 区块链核心实现
│   │   ├── state/     # 状态管理
│   │   └── txpool/    # 交易池管理
│   ├── types/         # 数据类型定义
│   │   ├── block/     # 区块相关类型
│   │   ├── tx/        # 交易相关类型
│   │   └── state/     # 状态相关类型
│   └── crypto/        # 加密相关功能
│       ├── hash/      # 哈希函数实现
│       └── merkle/    # Merkle树实现
├── pkg/               # 可导出的包
│   └── utils/        # 通用工具函数
├── scripts/          # 部署和管理脚本
│   ├── deploy.sh     # 部署脚本
│   └── test.sh       # 测试脚本
├── test/             # 测试文件和脚本
│   ├── integration/  # 集成测试
│   └── unit/        # 单元测试
├── docs/             # 项目文档
│   ├── api/         # API文档
│   └── design/      # 设计文档
└── .gitignore       # Git忽略文件配置
```

## 代码规范

### 1. 目录结构规范
- 遵循标准的Go项目布局
- 使用internal目录隔离内部包
- 使用pkg目录存放可导出包
- 测试代码与源码分离

### 2. 包设计原则
- 单一职责：每个包专注于单一功能
- 内聚性：相关功能放在同一包中
- 最小依赖：减少包之间的相互依赖
- 避免循环依赖

### 3. 命名规范
- 包名：使用小写单词，简短有意义
- 文件名：使用小写，下划线分隔
- 接口名：动词+er（如 Reader, Writer）
- 变量名：驼峰命名，简短清晰
- 常量名：全大写，下划线分隔

### 4. 错误处理
- 使用自定义错误类型
- 错误信息清晰明确
- 错误应该包含上下文信息
- 避免panic，使用error返回

### 5. 并发处理
- 明确锁的范围，最小化锁定时间
- 优先使用通道而不是共享内存
- 使用context控制goroutine生命周期
- 避免goroutine泄漏

### 6. 注释规范
- 包级别注释：描述包的用途和功能
- 函数注释：描述功能、参数和返回值
- 复杂逻辑注释：解释实现原理
- 避免无意义的注释

### 7. 测试规范
- 单元测试覆盖核心功能
- 使用表驱动测试
- 基准测试关键组件
- 集成测试验证系统功能

### 8. 日志规范
- 统一日志格式
- 合适的日志级别
- 包含必要的上下文信息
- 避免敏感信息泄露

### 9. 配置管理
- 使用环境变量
- 配置文件分环境
- 敏感信息加密存储
- 支持配置热重载

### 10. 版本控制
- 语义化版本号
- 清晰的提交信息
- 使用分支管理功能
- 及时处理代码冲突

### Git忽略配置

项目包含了 `.gitignore` 文件，用于排除编译生成的文件：

- 可执行文件（如 `zkrollup`、`*.exe`、`*.dll`、`*.so`、`*.dylib`）
- 测试二进制文件（`*.test`）
- 测试覆盖率文件（`*.out`）

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
- 交易签名系统
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

测试输出使用彩色标记：
- 绿色：测试步骤和成功信息
- 红色：错误信息
- 默认：测试结果和详细数据

测试完成后会自动清理进程，确保系统恢复到初始状态。

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
  ```

#### 查询交易
- **GET** `/api/v1/transaction/get?hash={transaction_hash}`
- **响应**:
  ```json
  {
    "hash": "hex_string",
    "from": "address",
    "to": "address",
    "value": 100,
    "nonce": 1,
    "status": "confirmed",
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
  ```

### 账户相关接口

#### 查询余额
- **GET** `/api/v1/balance/get?address={address}`
- **响应**:
  ```json
  {
    "address": "address",
    "balance": 1000
  }
  ```

#### 查询Nonce
- **GET** `/api/v1/account/nonce?address={address}`
- **响应**:
  ```json
  {
    "address": "address",
    "nonce": 1
  }
  ```

### 状态相关接口

#### 查询状态根
- **GET** `/api/v1/state/root`
- **响应**:
  ```json
  {
    "stateRoot": "hex_string"
  }
  ```

详细的API文档请参考 [API文档](docs/api.md)