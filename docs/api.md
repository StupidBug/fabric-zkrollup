# ZK Rollup API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints

### Transaction Operations

#### Send Transaction
Send a new transaction to the network.

- **URL**: `/transaction/send`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body**:
```json
{
    "from": "string",     // Sender address (hex string)
    "to": "string",       // Receiver address (hex string)
    "value": "string",    // Amount to transfer (decimal string)
    "nonce": "string"     // Transaction nonce (decimal string)
}
```

**Response**:
```json
{
    "hash": "string",      // Transaction hash
    "from": "string",      // Sender address
    "to": "string",        // Receiver address
    "value": "string",     // Transaction amount
    "nonce": "number",     // Transaction nonce
    "status": "string",    // Transaction status ("pending" or "confirmed")
    "timestamp": "number"  // Transaction timestamp
}
```

**Error Response**:
```json
{
    "error": "string"     // Error message
}
```

#### Get Transaction
Retrieve transaction details by hash.

- **URL**: `/transaction/get`
- **Method**: `GET`
- **Query Parameters**: `hash=[string]`

**Response**:
```json
{
    "hash": "string",      // Transaction hash
    "from": "string",      // Sender address
    "to": "string",        // Receiver address
    "value": "string",     // Transaction amount
    "nonce": "number",     // Transaction nonce
    "status": "string",    // Transaction status
    "timestamp": "number"  // Transaction timestamp
}
```

### Balance Operations

#### Get Balance
Get the balance of an address.

- **URL**: `/balance/get`
- **Method**: `GET`
- **Query Parameters**: `address=[string]`

**Response**:
```json
{
    "address": "string",   // Account address
    "balance": "string"    // Account balance (decimal string)
}
```

#### Set Balance (Testing Only)
Set the balance for an address (for testing purposes).

- **URL**: `/balance/set`
- **Method**: `POST`
- **Content-Type**: `application/json`

**Request Body**:
```json
{
    "address": "string",   // Account address
    "balance": "string"    // New balance (decimal string)
}
```

**Response**:
```json
{
    "address": "string",   // Account address
    "balance": "string"    // Updated balance
}
```

### State Operations

#### Get State Root
Get the current state root hash.

- **URL**: `/state/root`
- **Method**: `GET`

**Response**:
```json
{
    "stateRoot": "string"  // Current state root hash (hex string)
}
```

## Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Error Handling

All error responses follow this format:
```json
{
    "error": "string"     // Human-readable error message
}
```

Common error messages include:
- "Invalid address"
- "Invalid balance"
- "Invalid nonce"
- "Invalid hash format"
- "Transaction not found"
- "Insufficient balance"

## Examples

### Send Transaction Example

Request:
```bash
curl -X POST http://localhost:8080/api/v1/transaction/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": "100000",
    "nonce": "0"
  }'
```

Response:
```json
{
    "hash": "c772d6b215d3",
    "from": "0000000000000000000000000000000000000001",
    "to": "0000000000000000000000000000000000000002",
    "value": "100000",
    "nonce": 0,
    "status": "pending",
    "timestamp": 1647123456
}
```

### Get Balance Example

Request:
```bash
curl http://localhost:8080/api/v1/balance/get?address=0000000000000000000000000000000000000001
```

Response:
```json
{
    "address": "0000000000000000000000000000000000000001",
    "balance": "900000"
}
```

### Get State Root Example

Request:
```bash
curl http://localhost:8080/api/v1/state/root
```

Response:
```json
{
    "stateRoot": "a9ed704ad970d3799dad77f2fd67bef38b6489fc683a715e2c92fe6ef5297f4f"
}
``` 