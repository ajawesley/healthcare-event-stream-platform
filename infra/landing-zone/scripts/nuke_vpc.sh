#!/usr/bin/env bash
set -euo pipefail

VPC_ID="$1"

echo "🔍 Nuking VPC: $VPC_ID"

# Delete VPC Endpoints
echo "➡️ Deleting VPC Endpoints..."
aws ec2 describe-vpc-endpoints --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "VpcEndpoints[].VpcEndpointId" --output text | tr '\t' '\n' | \
  xargs -r -I {} aws ec2 delete-vpc-endpoint --vpc-endpoint-id {}

# Delete NAT Gateways
echo "➡️ Deleting NAT Gateways..."
NGWS=$(aws ec2 describe-nat-gateways --filter "Name=vpc-id,Values=$VPC_ID" \
  --query "NatGateways[].NatGatewayId" --output text)

for NGW in $NGWS; do
  echo "  Deleting NAT Gateway $NGW..."
  aws ec2 delete-nat-gateway --nat-gateway-id "$NGW"
  echo "  Waiting for NAT Gateway $NGW to delete..."
  aws ec2 wait nat-gateway-deleted --nat-gateway-ids "$NGW"
done

# Release Elastic IPs attached to ENIs in this VPC
echo "➡️ Releasing Elastic IPs for this VPC..."
ENIS=$(aws ec2 describe-network-interfaces --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "NetworkInterfaces[].NetworkInterfaceId" --output text)

for ENI in $ENIS; do
  ALLOC_ID=$(aws ec2 describe-addresses \
    --filters "Name=network-interface-id,Values=$ENI" \
    --query "Addresses[].AllocationId" --output text)

  if [[ "$ALLOC_ID" != "None" && -n "$ALLOC_ID" ]]; then
    echo "  Releasing EIP $ALLOC_ID attached to ENI $ENI..."
    aws ec2 release-address --allocation-id "$ALLOC_ID" || true
  fi
done

# Detach and delete Internet Gateways
echo "➡️ Deleting Internet Gateways..."
IGWS=$(aws ec2 describe-internet-gateways --filters "Name=attachment.vpc-id,Values=$VPC_ID" \
  --query "InternetGateways[].InternetGatewayId" --output text)

for IGW in $IGWS; do
  echo "  Detaching and deleting IGW $IGW..."
  aws ec2 detach-internet-gateway --internet-gateway-id "$IGW" --vpc-id "$VPC_ID"
  aws ec2 delete-internet-gateway --internet-gateway-id "$IGW"
done

# Delete Subnets
echo "➡️ Deleting Subnets..."
aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "Subnets[].SubnetId" --output text | tr '\t' '\n' | \
  xargs -r -I {} aws ec2 delete-subnet --subnet-id {}

# Delete Route Tables (non-main)
echo "➡️ Deleting Route Tables..."
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "RouteTables[?Associations[0].Main==false].RouteTableId" --output text | tr '\t' '\n' | \
  xargs -r -I {} aws ec2 delete-route-table --route-table-id {}

# Delete Security Groups (non-default) — WINDOWS-SAFE VERSION
echo "➡️ Deleting Security Groups..."
SGS=$(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "SecurityGroups[?GroupName!='default'].GroupId" --output text)

for SG in $SGS; do
  CLEANED_SG=$(echo "$SG" | tr -d '\r' | tr -d '\n' | xargs)
  if [[ "$CLEANED_SG" == sg-* ]]; then
    echo "  Deleting SG $CLEANED_SG..."
    aws ec2 delete-security-group --group-id "$CLEANED_SG" || true
  else
    echo "  ⚠️ Skipping malformed SG ID: '$SG'"
  fi
done

# Delete Network ACLs (non-default)
echo "➡️ Deleting Network ACLs..."
aws ec2 describe-network-acls --filters "Name=vpc-id,Values=$VPC_ID" \
  --query "NetworkAcls[?IsDefault==false].NetworkAclId" --output text | tr '\t' '\n' | \
  xargs -r -I {} aws ec2 delete-network-acl --network-acl-id {}

# Delete VPC
echo "➡️ Deleting VPC..."
aws ec2 delete-vpc --vpc-id "$VPC_ID"

echo "🎉 VPC $VPC_ID successfully nuked."
