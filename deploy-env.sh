#!/bin/bash

# Configuration
EC2_USER="ubuntu"
EC2_HOST="ec2-3-134-228-165.us-east-2.compute.amazonaws.com"
EC2_KEY="keys/ec2.pem"
SERVER_DIR="~/server"
LOCAL_ENV_FILE=".env"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🚀 Starting environment deployment..."

# Check if key file exists and set permissions
if [ ! -f "$EC2_KEY" ]; then
    echo -e "${RED}Error: SSH key file not found: $EC2_KEY${NC}"
    exit 1
else
    echo "🔒 Setting correct permissions for EC2 key..."
    chmod 600 "$EC2_KEY"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Error: Failed to set permissions on key file${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Key file permissions set successfully${NC}"
fi

# Check if env file exists
if [ ! -f "$LOCAL_ENV_FILE" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    exit 1
fi

# Create server directory if it doesn't exist
ssh -i "$EC2_KEY" "$EC2_USER@$EC2_HOST" "mkdir -p $SERVER_DIR"

# Transfer the env file using scp
echo "📦 Transferring environment files..."
scp -i "$EC2_KEY" "$LOCAL_ENV_FILE" "$EC2_USER@$EC2_HOST:$SERVER_DIR/.env"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Environment files deployed successfully!${NC}"
    echo -e "${GREEN}🔒 Securing environment files...${NC}"
    
    # Set correct permissions on the server
    ssh -i "$EC2_KEY" "$EC2_USER@$EC2_HOST" "chmod 600 $SERVER_DIR/.env"
    
    echo -e "${GREEN}✨ Deployment complete!${NC}"
else
    echo -e "${RED}❌ Deployment failed${NC}"
    exit 1
fi
