#!/usr/bin/env bash
set -euo pipefail

ORG_ACCOUNT_ID=$1
DEV_ACCOUNT_ID=$2
REGION=$3
APP=$4
ENV=$5

echo "=============================================="
echo "🔍 VALIDATING LANDING ZONE OIDC + CONFIG SETUP"
echo "=============================================="

echo ""
echo "1️⃣ Checking ORG-LEVEL OIDC Provider..."
aws iam get-open-id-connect-provider \
  --open-id-connect-provider-arn arn:aws:iam::$ORG_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com \
  >/dev/null && echo "✔ OIDC provider exists" || echo "❌ MISSING OIDC PROVIDER"

echo ""
echo "2️⃣ Checking ORG-LEVEL GitHub Deploy Role..."
aws iam get-role \
  --role-name ${APP}-github-oidc-deploy-role \
  >/dev/null && echo "✔ Deploy role exists" || echo "❌ MISSING DEPLOY ROLE"

echo ""
echo "3️⃣ Checking Deploy Role Trust Policy..."
aws iam get-role \
  --role-name ${APP}-github-oidc-deploy-role \
  --query "Role.AssumeRolePolicyDocument.Statement" \
  --output json

echo ""
echo "4️⃣ Checking Deploy Role Inline Policy..."
aws iam list-role-policies \
  --role-name ${APP}-github-oidc-deploy-role \
  --query "PolicyNames" \
  --output json

echo ""
echo "5️⃣ Checking Deploy Role PassRole Permissions..."
aws iam get-role-policy \
  --role-name ${APP}-github-oidc-deploy-role \
  --policy-name ${APP}-github-deploy-policy \
  --query "PolicyDocument.Statement[?Action=='iam:PassRole']" \
  --output json

echo ""
echo "6️⃣ Checking Workload Config Recorder Role..."
aws iam get-role \
  --role-name ${APP}-${ENV}-config-recorder-role \
  --output json >/dev/null && echo "✔ Config Recorder role exists" || echo "❌ MISSING CONFIG RECORDER ROLE"

echo ""
echo "7️⃣ Checking Config Recorder Trust Policy..."
aws iam get-role \
  --role-name ${APP}-${ENV}-config-recorder-role \
  --query "Role.AssumeRolePolicyDocument.Statement" \
  --output json

echo ""
echo "8️⃣ Checking AWS Config Recorder Status..."
aws configservice describe-configuration-recorders \
  --region $REGION \
  --query "ConfigurationRecorders" \
  --output json

echo ""
echo "9️⃣ Checking AWS Config Recorder Status Details..."
aws configservice describe-configuration-recorder-status \
  --region $REGION \
  --query "ConfigurationRecorderStatuses" \
  --output json

echo ""
echo "🔟 Checking AWS Config Delivery Channel..."
aws configservice describe-delivery-channels \
  --region $REGION \
  --query "DeliveryChannels" \
  --output json

echo ""
echo "1️⃣1️⃣ Checking CloudTrail Integration..."
aws cloudtrail describe-trails \
  --region $REGION \
  --query "trailList" \
  --output json

echo ""
echo "1️⃣2️⃣ Checking GuardDuty Detector..."
aws guardduty list-detectors \
  --region $REGION \
  --query "DetectorIds" \
  --output json

echo ""
echo "=============================================="
echo "🏁 VALIDATION COMPLETE"
echo "=============================================="


