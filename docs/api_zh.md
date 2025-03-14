# ZK Rollup API 文档

## 基础URL
```
http://localhost:8080/api/v1
```

## 签名生成和验证

### 签名生成步骤
1. 计算交易哈希：
   - 按顺序拼接以下字段：
     * From 地址
     * To 地址
     * Value（转账金额的字节表示）
     * Nonce（交易序号）
     * Timestamp（时间戳）
   - 对拼接后的数据计算 SHA-256 哈希

2. 使用私钥对哈希进行签名：
   - 使用 ECDSA 算法进行签名
   - 生成 R、S 两个签名分量

### 签名验证
服务器在接收到交易时会进行以下验证：
1. 检查签名是否完整（R、S 值是否存在）
2. 验证签名值的范围（R、S 必须为正数）
3. 使用发送方的公钥验证签名

### 测试工具
为方便测试，我们提供了一个简单的签名生成工具：

```go
package main

import (
    "crypto/ecdsa"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "math/big"
)

func main() {
    // 生成私钥
    privateKey, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
    
    // 交易数据
    txHash := []byte("your-transaction-hash")
    
    // 签名
    r, s, _ := ecdsa.Sign(rand.Reader, privateKey, txHash)
    
    // 输出签名
    fmt.Printf("R: %x\n", r)
    fmt.Printf("S: %x\n", s)
}
```

## 接口列表

### 交易相关接口

#### 发送交易
发送新的交易到网络。每笔交易必须包含有效的签名。

- **接口**: `/transaction/send`
- **方法**: `POST`
- **Content-Type**: `application/json`

**请求体**:
```json
{
    "from": "string",     // 发送方地址（16进制字符串）
    "to": "string",       // 接收方地址（16进制字符串）
    "value": "string",    // 转账金额（十进制字符串）
    "nonce": "string",    // 交易序号（十进制字符串）
    "signature": {        // 交易签名
        "r": "string",    // 签名 R 值（16进制字符串）
        "s": "string"      // 签名 S 值（16进制字符串）
    }
}
```

**签名数据结构**:
交易签名需要对以下字段进行签名（按顺序）：
1. from（地址）
2. to（地址）
3. value（金额）
4. nonce（交易序号）

签名过程：
1. 将上述字段按顺序拼接
2. 计算拼接后数据的 Keccak-256 哈希
3. 使用发送方的私钥对哈希进行签名
4. 生成签名的 r、s 值

**响应**:
```json
{
    "hash": "string",      // 交易哈希
    "from": "string",      // 发送方地址
    "to": "string",        // 接收方地址
    "value": "string",     // 交易金额
    "nonce": "number",     // 交易序号
    "status": "string",    // 交易状态（"pending"或"confirmed"）
    "timestamp": "number", // 交易时间戳
    "signature": {         // 交易签名
        "r": "string",     // 签名 R 值
        "s": "string"      // 签名 S 值
    }
}
```

**错误响应**:
```json
{
    "error": "string"     // 错误信息
}
```

**可能的错误信息**:
- "无效的签名 R 值"
- "无效的签名 S 值"
- "缺少签名"
- "签名验证失败"
- "签名值超出有效范围"

#### 获取交易详情
通过交易哈希获取交易详情。

- **接口**: `/transaction/get`
- **方法**: `GET`
- **查询参数**: `hash=[string]`

**响应**:
```json
{
    "hash": "string",      // 交易哈希
    "from": "string",      // 发送方地址
    "to": "string",        // 接收方地址
    "value": "string",     // 交易金额
    "nonce": "number",     // 交易序号
    "status": "string",    // 交易状态
    "timestamp": "number"  // 交易时间戳
}
```

### 余额相关接口

#### 查询余额
获取指定地址的余额。

- **接口**: `/balance/get`
- **方法**: `GET`
- **查询参数**: `address=[string]`

**响应**:
```json
{
    "address": "string",   // 账户地址
    "balance": "string"    // 账户余额（十进制字符串）
}
```

#### 设置余额（仅用于测试）
设置指定地址的余额（仅用于测试环境）。

- **接口**: `/balance/set`
- **方法**: `POST`
- **Content-Type**: `application/json`

**请求体**:
```json
{
    "address": "string",   // 账户地址
    "balance": "string"    // 新的余额（十进制字符串）
}
```

**响应**:
```json
{
    "address": "string",   // 账户地址
    "balance": "string"    // 更新后的余额
}
```

### 状态相关接口

#### 获取状态根哈希
获取当前的状态根哈希值。

- **接口**: `/state/root`
- **方法**: `GET`

**响应**:
```json
{
    "stateRoot": "string"  // 当前状态根哈希（16进制字符串）
}
```

## 状态码说明

- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数无效
- `404 Not Found`: 资源未找到
- `500 Internal Server Error`: 服务器内部错误

## 错误处理

所有错误响应都遵循以下格式：
```json
{
    "error": "string"     // 人类可读的错误信息
}
```

常见错误信息包括：
- "无效的地址格式"
- "无效的余额数值"
- "无效的交易序号"
- "无效的哈希格式"
- "交易未找到"
- "余额不足"
- "签名验证失败"
- "无效的签名格式"
- "签名与发送方地址不匹配"

## 使用示例

### 发送交易示例

请求：
```bash
curl -X POST http://localhost:8080/api/v1/transaction/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": "100000",
    "nonce": "0",
    "signature": {
        "r": "28a9be0f3646b5934fb068c8073ad1ca5def791dd016d36c44ce3ab5ed9d8e2a",
        "s": "67b4c3aa14c96da9aaf0e965f51f7a7dce1182835dc0e1ee264707e2a6b34f48"
    }
  }'
```

响应：
```json
{
    "hash": "c772d6b215d3",
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": "100000",
    "nonce": 0,
    "status": "pending",
    "timestamp": 1647123456,
    "signature": {
        "r": "28a9be0f3646b5934fb068c8073ad1ca5def791dd016d36c44ce3ab5ed9d8e2a",
        "s": "67b4c3aa14c96da9aaf0e965f51f7a7dce1182835dc0e1ee264707e2a6b34f48"
    }
}
```

### 查询余额示例

请求：
```bash
curl http://localhost:8080/api/v1/balance/get?address=0000000000000000000000000000000000000001
```

响应：
```json
{
    "address": "0000000000000000000000000000000000000001",
    "balance": "900000"
}
```

### 获取状态根示例

请求：
```bash
curl http://localhost:8080/api/v1/state/root
```

响应：
```json
{
    "stateRoot": "a9ed704ad970d3799dad77f2fd67bef38b6489fc683a715e2c92fe6ef5297f4f"
}
```

## 注意事项

1. 所有地址必须是40个字符的16进制字符串（不包含"0x"前缀）
2. 金额和余额使用十进制字符串表示，避免精度损失
3. 交易序号(nonce)必须从0开始递增
4. 状态根哈希是64个字符的16进制字符串
5. 时间戳使用Unix时间戳（秒）
6. 签名相关要求：
   - 签名必须使用 ECDSA 算法
   - r、s 值必须是65个字节的16进制字符串
   - v 值用于恢复公钥，通常是 0x1b 或 0x1c
   - 每笔交易的签名必须与发送方地址匹配
   - 签名验证失败的交易会被立即拒绝 