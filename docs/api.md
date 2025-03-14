# ZK Rollup Sidechain API 文档

## 基本信息

- 基础URL: `http://localhost:8080`
- 所有请求和响应均使用 JSON 格式
- 所有地址字段均为 20 字节的十六进制字符串（不含 0x 前缀）
- 所有金额字段均为整数类型（int）
- 所有签名字段均为 65 字节的十六进制字符串（不含 0x 前缀）
- 所有公钥字段均为 64 字节的十六进制字符串（不含 0x 前缀）

## API 端点

### 1. 发送交易

**请求**:
```
POST /api/v1/transaction/send
```

**请求体**:
```json
{
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": 100,
    "nonce": 1,
    "signature": "...", // 65字节的ECDSA签名
    "public_key": "..." // 64字节的公钥
}
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "hash": "...", // 32字节的交易哈希
        "status": "pending" // pending, confirmed, failed
    }
}
```

**错误响应**:
```json
{
    "status": "error",
    "error": {
        "code": "invalid_signature",
        "message": "Invalid transaction signature"
    }
}
```

常见错误:
- `invalid_signature`: 无效的交易签名
- `invalid_nonce`: 无效的 nonce 值
- `insufficient_balance`: 余额不足
- `invalid_address`: 无效的地址格式
- `invalid_value`: 无效的转账金额

### 2. 查询交易

**请求**:
```
GET /api/v1/transaction/get?hash={transaction_hash}
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "hash": "...",
        "from": "0000000000000000000000000000000000000001",
        "to": "0000000000000000000000000000000000000002",
        "value": 100,
        "nonce": 1,
        "status": "confirmed",
        "signature": "...",
        "public_key": "..."
    }
}
```

### 3. 查询交易池

**请求**:
```
GET /api/v1/transactions/pool
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "transactions": [
            {
                "hash": "...",
                "from": "0000000000000000000000000000000000000001",
                "to": "0000000000000000000000000000000000000002",
                "value": 100,
                "nonce": 1,
                "status": "pending",
                "signature": "...",
                "public_key": "..."
            }
        ]
    }
}
```

### 4. 查询余额

**请求**:
```
GET /api/v1/balance/get?address={address}
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "address": "0000000000000000000000000000000000000001",
        "balance": 1000
    }
}
```

### 5. 查询账户 Nonce

**请求**:
```
GET /api/v1/account/nonce?address={address}
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "address": "0000000000000000000000000000000000000001",
        "nonce": 1
    }
}
```

### 6. 查询状态根

**请求**:
```
GET /api/v1/state/root
```

**响应**:
```json
{
    "status": "success",
    "data": {
        "root": "..." // 32字节的状态根哈希
    }
}
```

## 状态码

- 200: 请求成功
- 400: 请求参数错误
- 401: 未授权
- 404: 资源不存在
- 500: 服务器内部错误

## 错误处理

所有错误响应均使用以下格式:

```json
{
    "status": "error",
    "error": {
        "code": "error_code",
        "message": "详细错误信息"
    }
}
```

常见错误代码:
- `invalid_request`: 无效的请求格式
- `invalid_parameter`: 无效的参数
- `not_found`: 资源不存在
- `internal_error`: 服务器内部错误

## 安全性说明

1. 所有交易必须包含有效的 ECDSA 签名
2. 签名必须使用发送方地址对应的私钥生成
3. 交易的 nonce 必须严格递增
4. 所有金额必须为非负整数
5. 所有地址必须为有效的 20 字节十六进制字符串

## 最佳实践

1. 发送交易前先查询当前 nonce
2. 使用 websocket 监听交易状态变化
3. 定期查询交易状态直到确认
4. 保持私钥安全，不要在请求中传输
5. 验证所有响应数据的完整性 