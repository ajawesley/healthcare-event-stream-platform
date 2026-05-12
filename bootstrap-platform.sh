#!/usr/bin/env bash
set -euo pipefail

############################################
# Bootstrap Script for Landing Zone + App Platform
############################################

APP_NAME="hesp"
ENV="${1:-dev}"
AWS_REGION="us-east-1"

echo "Bootstrapping environment: $ENV"
echo "AWS Region: $AWS_REGION"

############################################
# Detect Current AWS Account
############################################

AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "Current AWS Account ID: $AWS_ACCOUNT_ID"

############################################
# Detect Workloads Account ID (Landing Zone Phase 2)
############################################

WORKLOADS_ACCOUNT_ID=$(aws organizations list-accounts \
  --query "Accounts[?Name=='Workloads'].Id" --output text)

if [[ -z "$WORKLOADS_ACCOUNT_ID" ]]; then
  echo "ERROR: Could not detect Workloads account ID from AWS Organizations"
  exit 1
fi

echo "Workloads Account ID: $WORKLOADS_ACCOUNT_ID"

############################################
# Export Environment Variables
############################################

export APP_NAME
export ENV
export AWS_REGION
export AWS_ACCOUNT_ID
export WORKLOADS_ACCOUNT_ID

echo "Environment variables exported."

############################################
# Initialize Landing Zone (Phase 1)
############################################

echo "Initializing Landing Zone Phase 1 (Org)..."
(
  cd infra/landing-zone/org
  terraform init
)

############################################
# Initialize Landing Zone (Phase 2)
############################################

echo "Initializing Landing Zone Phase 2 (Workloads)..."
(
  cd infra/landing-zone/workloads
  terraform init
)

############################################
# Initialize Application Environment
############################################

echo "Initializing Application Environment: $ENV..."
(
  cd infra/envs/$ENV
  terraform init
)

############################################
# Summary
############################################

echo ""
echo "Bootstrap complete."
echo "You can now run:"
echo "  make lz-deploy"
echo "  make deploy-all ENV=$ENV"
echo ""

