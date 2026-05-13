#!/usr/bin/env bash
set -euo pipefail

ORG_ACCOUNT_ID=$1
DEV_ACCOUNT_ID=$2
REGION=$3
APP=$4
ENV=$5

echo "===================================================="
echo " LANDING ZONE POST-DEPLOYMENT VALIDATION"
echo "===================================================="

###############################################
# 1. VPC VALIDATION
###############################################
echo ""
echo "1) Checking VPC..."
aws ec2 describe-vpcs \
  --filters Name=tag:Name,Values="${APP}-${ENV}-vpc" \
  --query "Vpcs[].{VpcId:VpcId,CIDR:CidrBlock,State:State}" \
  --output table

echo ""
echo "2) Checking Subnets..."
aws ec2 describe-subnets \
  --filters Name=vpc-id,Values=$(aws ec2 describe-vpcs \
    --filters Name=tag:Name,Values="${APP}-${ENV}-vpc" \
    --query "Vpcs[0].VpcId" --output text) \
  --query "Subnets[].{Id:SubnetId,CIDR:CidrBlock,AZ:AvailabilityZone,Tags:Tags}" \
  --output table

echo ""
echo "3) Checking Route Tables..."
aws ec2 describe-route-tables \
  --filters Name=vpc-id,Values=$(aws ec2 describe-vpcs \
    --filters Name=tag:Name,Values="${APP}-${ENV}-vpc" \
    --query "Vpcs[0].VpcId" --output text) \
  --query "RouteTables[].{Id:RouteTableId,Routes:Routes}" \
  --output table

echo ""
echo "4) Checking VPC Endpoints..."
aws ec2 describe-vpc-endpoints \
  --filters Name=vpc-id,Values=$(aws ec2 describe-vpcs \
    --filters Name=tag:Name,Values="${APP}-${ENV}-vpc" \
    --query "Vpcs[0].VpcId" --output text) \
  --query "VpcEndpoints[].{Id:VpcEndpointId,Service:ServiceName,State:State}" \
  --output table

###############################################
# 2. SECURITY SERVICES
###############################################
echo ""
echo "5) Checking GuardDuty Detector..."
aws guardduty list-detectors \
  --region "$REGION" \
  --query "DetectorIds" \
  --output table

echo ""
echo "6) Checking GuardDuty Member Enrollment..."
DETECTOR=$(aws guardduty list-detectors --region "$REGION" --query DetectorIds[0] --output text)
aws guardduty list-members \
  --detector-id "$DETECTOR" \
  --region "$REGION" \
  --query "Members[].{Account:AccountId,Status:RelationshipStatus}" \
  --output table

echo ""
echo "7) Checking CloudTrail Org Trail..."
aws cloudtrail describe-trails \
  --region "$REGION" \
  --query "trailList[].{Name:Name,IsMultiRegion:IsMultiRegionTrail,HomeRegion:HomeRegion}" \
  --output table

###############################################
# 3. AWS CONFIG
###############################################
echo ""
echo "8) Checking Config Recorder..."
aws configservice describe-configuration-recorders \
  --region "$REGION" \
  --query "ConfigurationRecorders" \
  --output table

echo ""
echo "9) Checking Config Recorder Status..."
aws configservice describe-configuration-recorder-status \
  --region "$REGION" \
  --query "ConfigurationRecorderStatuses" \
  --output table

echo ""
echo "10) Checking Config Delivery Channel..."
aws configservice describe-delivery-channels \
  --region "$REGION" \
  --query "DeliveryChannels" \
  --output table

echo ""
echo "11) Checking Config Aggregator..."
aws configservice describe-configuration-aggregators \
  --region "$REGION" \
  --query "ConfigurationAggregators[].{Name:ConfigurationAggregatorName,AccountAggregationSources:AccountAggregationSources}" \
  --output table

###############################################
# 4. IAM + SCP VALIDATION
###############################################
echo ""
echo "12) Checking SCPs Attached to ORG Root..."
aws organizations list-policies-for-target \
  --target-id "$ORG_ACCOUNT_ID" \
  --filter SERVICE_CONTROL_POLICY \
  --query "Policies[].{Name:Name,Id:Id}" \
  --output table

echo ""
echo "13) Checking IAM Boundary Policies..."
aws iam list-policies \
  --scope Local \
  --query "Policies[?contains(PolicyName, 'boundary')].{Name:PolicyName,Arn:Arn}" \
  --output table

###############################################
# 5. CROSS-ACCOUNT TRUST
###############################################
echo ""
echo "14) Checking Cross-Account AssumeRole Permissions..."
aws iam get-role \
  --role-name "${APP}-${ENV}-workload-role" \
  --query "Role.AssumeRolePolicyDocument.Statement" \
  --output json 2>/dev/null || echo "❌ Workload role trust policy missing"

echo ""
echo "===================================================="
echo " VALIDATION COMPLETE"
echo "===================================================="

