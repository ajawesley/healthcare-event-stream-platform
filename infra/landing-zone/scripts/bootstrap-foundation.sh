#!/usr/bin/env bash
set -euo pipefail

AWS_REGION="us-east-1"
BUCKET="hesp-landing-zone-tfstate"
TABLE="hesp-landing-zone-locks"

echo "Creating S3 backend bucket..."

if [ "$AWS_REGION" = "us-east-1" ]; then
  # us-east-1 does NOT allow LocationConstraint
  aws s3api create-bucket \
    --bucket "$BUCKET" \
    --region "$AWS_REGION" || true
else
  # all other regions require LocationConstraint
  aws s3api create-bucket \
    --bucket "$BUCKET" \
    --region "$AWS_REGION" \
    --create-bucket-configuration LocationConstraint="$AWS_REGION" || true
fi

echo "Enabling versioning..."
aws s3api put-bucket-versioning \
  --bucket "$BUCKET" \
  --versioning-configuration Status=Enabled || true

echo "Creating DynamoDB lock table..."
aws dynamodb create-table \
  --table-name "$TABLE" \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST || true

echo "Bootstrap complete."
