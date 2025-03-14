#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080/api/v1"


curl -s "$BASE_URL/state/root" | jq