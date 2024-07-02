#!/bin/bash
install_jq() {
    if ! command -v jq &> /dev/null; then
        echo "jq is not installed. Installing..."
        if [[ "$(uname)" == "Darwin" ]]; then
            # macOS
            brew install jq
        elif [[ "$(uname)" == "Linux" ]]; then
            # Linux
            sudo apt-get update
            sudo apt-get install jq -y
        else
            echo "Unsupported OS. Please install jq manually: https://stedolan.github.io/jq/download/"
            exit 1
        fi
    fi
}

# Check and install jq
install_jq
# API endpoint URL
BASE_URL="http://localhost/api/v1/accounts"

# Read data from accounts.json and iterate over each object
cat account.json | jq -c '.[]' | while read -r payload; do
    echo "Sending payload: $payload"
    
    # Make a POST request using curl
    curl -X POST \
         -H "Content-Type: application/json" \
         -d "$payload" \
         "$BASE_URL"
    
    echo "-----------------------"
done
