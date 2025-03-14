#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
NC='\033[0m' # No Color
RED='\033[0;31m'

# Base URL
BASE_URL="http://localhost:8080/api/v1"

# Function to check if server is running
check_server() {
    echo -e "${GREEN}Checking if server is running...${NC}"
    if ! curl -s "$BASE_URL/state/root" > /dev/null; then
        echo -e "${RED}Server is not running. Please start the server first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}Server is running.${NC}"
}

# Function to get state root
get_state_root() {
    echo -e "${GREEN}Getting state root...${NC}"
    curl -s "$BASE_URL/state/root" | jq -r '.stateRoot'
}

# Function to get balance
get_balance() {
    local address=$1
    echo -e "${GREEN}Getting balance for address $address...${NC}"
    curl -s "$BASE_URL/balance/get?address=$address" | jq -r '.balance'
}

# Function to get transaction status with debug info
get_tx_status() {
    local tx_hash=$1
    echo "Debug - Requesting status for transaction: $tx_hash" >&2
    
    # Check if tx_hash is empty
    if [ -z "$tx_hash" ]; then
        echo "Debug - Empty transaction hash provided" >&2
        return 1
    fi
    
    # Get only the response body
    local response=$(curl -s "$BASE_URL/transaction/get?hash=$tx_hash")
    local curl_status=$?
    
    if [ $curl_status -ne 0 ]; then
        echo "Debug - Curl request failed with status $curl_status" >&2
        return 1
    fi
    
    echo "Debug - Full response for tx $tx_hash: $response" >&2
    
    # Check if response is empty
    if [ -z "$response" ]; then
        echo "Debug - Empty response from server" >&2
        return 1
    fi
    
    # Extract status using jq
    local status=$(echo "$response" | jq -r '.status // empty')
    
    if [ -z "$status" ]; then
        echo "Debug - No status field in response" >&2
        return 1
    fi
    
    echo "$status"
}

# Function to wait for transaction confirmation with retries
wait_for_confirmation() {
    local tx_hash=$1
    local max_retries=10
    local retry_count=0
    local status=""

    while [ $retry_count -lt $max_retries ]; do
        echo "Debug - Checking transaction status (attempt $((retry_count + 1))/$max_retries)" >&2
        status=$(get_tx_status "$tx_hash")
        echo "Debug - Transaction $tx_hash status: $status" >&2
        
        if [ "$status" = "1" ] || [ "$status" = "confirmed" ]; then
            echo "Debug - Transaction confirmed successfully" >&2
            return 0
        elif [ "$status" = "0" ] || [ "$status" = "pending" ]; then
            echo "Debug - Transaction still pending" >&2
        elif [ -z "$status" ]; then
            echo "Debug - No status returned from server" >&2
        else
            echo "Debug - Unknown status: $status" >&2
        fi
        
        retry_count=$((retry_count + 1))
        sleep 3
    done
    
    echo "Failed to confirm transaction after $max_retries attempts" >&2
    return 1
}

# Function to get nonce
get_nonce() {
    local address=$1
    curl -s "$BASE_URL/account/nonce?address=$address" | jq -r '.nonce'
}

# Function to send a transaction
send_transaction() {
    local from=$1
    local to=$2
    local value=$3
    local nonce=$4
    local priv_key=$5
    local sig_r=$6
    local sig_s=$7
    local pub_x=$8
    local pub_y=$9

    echo -e "${GREEN}Sending transaction...${NC}" >&2
    local tx_response=$(curl -s -X POST "$BASE_URL/transaction/send" \
        -H "Content-Type: application/json" \
        -d "{
            \"from\": \"$from\",
            \"to\": \"$to\",
            \"value\": \"$value\",
            \"nonce\": \"$nonce\",
            \"signature\": {
                \"r\": \"$sig_r\",
                \"s\": \"$sig_s\"
            },
            \"publicKey\": {
                \"x\": \"$pub_x\",
                \"y\": \"$pub_y\"
            }
        }")

    # Extract hash first
    local tx_hash=$(echo "$tx_response" | jq -r '.hash')
    
    # Print response for debugging
    echo "Transaction response: $tx_response" >&2
    
    # Return just the hash
    echo "$tx_hash"
}

# Main test process
echo -e "${GREEN}Starting transfer test...${NC}"

# Check if server is running
check_server

# Get initial state root
initial_root=$(get_state_root)
echo "Initial state root: $initial_root"

# Generate key pair for sender
echo -e "${GREEN}Generating sender key pair...${NC}"
sender_output=$(./keygen -genkey)
sender_priv_key=$(echo "$sender_output" | grep "Private key:" | cut -d' ' -f3)
sender_address="0000000000000000000000000000000000000001"

# Generate key pair for receiver
echo -e "${GREEN}Generating receiver key pair...${NC}"
receiver_output=$(./keygen -genkey)
receiver_address="0000000000000000000000000000000000000002"

# Get initial balances
echo -e "${GREEN}Initial balances:${NC}"
sender_initial_balance=$(curl -s "$BASE_URL/balance/get?address=$sender_address" | jq -r '.balance')
receiver_initial_balance=$(curl -s "$BASE_URL/balance/get?address=$receiver_address" | jq -r '.balance')

echo "Sender ($sender_address): $sender_initial_balance"
echo "Receiver ($receiver_address): $receiver_initial_balance"

# Array to store transaction hashes
declare -a tx_hashes

# Send multiple transactions
num_transactions=3
transfer_amount=100
total_transfer=0

for ((i=0; i<num_transactions; i++)); do
    echo -e "${GREEN}Processing transaction $((i+1))/$num_transactions...${NC}"
    
    # Get current nonce
    current_nonce=$(get_nonce $sender_address)
    if [ -z "$current_nonce" ] || [ "$current_nonce" = "null" ]; then
        current_nonce=0
    fi
    echo "Current nonce: $current_nonce"

    # Sign transaction
    echo -e "${GREEN}Signing transaction...${NC}"
    sign_output=$(./keygen -sign \
        -from $sender_address \
        -to $receiver_address \
        -value $transfer_amount \
        -nonce $current_nonce \
        -privkey $sender_priv_key)

    echo "Sign output: $sign_output"

    # Extract signature components and public key
    sig_r=$(echo "$sign_output" | grep "Signature R:" | cut -d' ' -f3)
    sig_s=$(echo "$sign_output" | grep "Signature S:" | cut -d' ' -f3)
    pub_x=$(echo "$sign_output" | grep "Public key X:" | cut -d' ' -f4)
    pub_y=$(echo "$sign_output" | grep "Public key Y:" | cut -d' ' -f4)

    # Send transaction
    tx_hash=$(send_transaction "$sender_address" "$receiver_address" "$transfer_amount" "$current_nonce" "$sender_priv_key" "$sig_r" "$sig_s" "$pub_x" "$pub_y")
    if [ -z "$tx_hash" ]; then
        echo -e "${RED}Failed to get transaction hash from response${NC}"
        exit 1
    fi
    tx_hashes[$i]=$tx_hash
    total_transfer=$((total_transfer + transfer_amount))

    # Wait for transaction to be confirmed
    echo -e "${GREEN}Waiting for transaction $((i+1)) to be confirmed...${NC}"
    if ! wait_for_confirmation "$tx_hash"; then
        echo -e "${RED}Test failed: Transaction $((i+1)) was not confirmed in time${NC}"
        exit 1
    fi
    echo -e "${GREEN}Transaction $((i+1)) confirmed successfully${NC}"

    # Get updated balances after each transaction
    echo -e "${GREEN}Checking balances after transaction $((i+1))...${NC}"
    current_sender_balance=$(curl -s "$BASE_URL/balance/get?address=$sender_address" | jq -r '.balance')
    current_receiver_balance=$(curl -s "$BASE_URL/balance/get?address=$receiver_address" | jq -r '.balance')
    echo "Current sender balance: $current_sender_balance"
    echo "Current receiver balance: $current_receiver_balance"

    # Add a small delay before next transaction
    sleep 2
done

# Get final balances
echo -e "${GREEN}Final balances:${NC}"
sender_final_balance=$(curl -s "$BASE_URL/balance/get?address=$sender_address" | jq -r '.balance')
receiver_final_balance=$(curl -s "$BASE_URL/balance/get?address=$receiver_address" | jq -r '.balance')

echo "Sender ($sender_address): $sender_final_balance"
echo "Receiver ($receiver_address): $receiver_final_balance"

# Get final state root
final_root=$(get_state_root)
echo "Final state root: $final_root"

# Verify state root has changed
if [ "$initial_root" != "$final_root" ]; then
    echo -e "${GREEN}Test passed: State root changed after transfers${NC}"
else
    echo -e "${RED}Test failed: State root did not change${NC}"
    exit 1
fi

# Convert values to integers for comparison
sender_final_balance_int=$((sender_final_balance))
receiver_final_balance_int=$((receiver_final_balance))
expected_sender_balance=$((sender_initial_balance - total_transfer))
expected_receiver_balance=$((receiver_initial_balance + total_transfer))

# Verify balances changed correctly
if [ "$sender_final_balance_int" -eq "$expected_sender_balance" ] && \
   [ "$receiver_final_balance_int" -eq "$expected_receiver_balance" ]; then
    echo -e "${GREEN}Test passed: Balances updated correctly${NC}"
else
    echo -e "${RED}Test failed: Balances not updated correctly${NC}"
    echo "Expected sender balance: $expected_sender_balance, got: $sender_final_balance_int"
    echo "Expected receiver balance: $expected_receiver_balance, got: $receiver_final_balance_int"
    exit 1
fi

echo -e "${GREEN}All tests passed successfully!${NC}"

# Get all blocks with formatted output
echo -e "\n${GREEN}Fetching all blocks...${NC}"
curl -s "$BASE_URL/blocks" | jq -r '
  "Blockchain Status:",
  "-----------------",
  "Total Blocks: \(.data.blocks | length)",
  "",
  (.data.blocks | to_entries | .[] | 
    "Block #\(.value.height)",
    "Hash: \(.value.hash)",
    "Previous Hash: \(.value.prevHash)",
    "Merkle Root: \(.value.merkleRoot)",
    "State Root: \(.value.stateRoot)",
    "Timestamp: \(.value.timestamp)",
    "Transaction Count: \(.value.transactionCount)",
    "",
    "Transactions:",
    (if (.value.transactions | length) > 0 then
      (.value.transactions | to_entries | .[] |
        "  Transaction #\(.key + 1)",
        "  Hash: \(.value.hash)",
        "  From: \(.value.from)",
        "  To: \(.value.to)",
        "  Value: \(.value.value)",
        "  Nonce: \(.value.nonce)",
        "  Status: \(.value.status)",
        "  Timestamp: \(.value.timestamp)",
        ""
      )
    else
      "  No transactions in this block"
    end),
    "-----------------"
  )
' 