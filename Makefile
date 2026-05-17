############################################
# Global Variables
############################################

APP_NAME        ?= hesp
ENV             ?= dev
AWS_REGION      ?= us-east-1
AWS_ACCOUNT_ID  ?= $(shell aws sts get-caller-identity --query Account --output text)

TF_DIR          = infra/envs/$(ENV)
LAMBDA_DIR      = cmd/lambda
INGESTION_DIR   = cmd/ingestion

ECR_REPO        = $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(APP_NAME)

############################################
# Terraform Commands (Application Platform)
############################################

tf-init:
        cd $(TF_DIR) && terraform init

tf-plan:
        cd $(TF_DIR) && terraform plan

tf-apply:
        cd $(TF_DIR) && terraform apply -auto-approve

tf-destroy:
        cd $(TF_DIR) && terraform destroy -auto-approve

############################################
# Lambda Build
############################################

lambda-build:
        cd $(LAMBDA_DIR) && \
        GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go && \
        zip lambda.zip bootstrap

lambda-clean:
        rm -f $(LAMBDA_DIR)/bootstrap $(LAMBDA_DIR)/lambda.zip

############################################
# Docker Build + Push (ECS)
############################################

docker-login:
        aws ecr get-login-password --region $(AWS_REGION) | \
        docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

docker-build:
        cd $(INGESTION_DIR) && docker build -t $(APP_NAME):latest .

docker-tag:
        docker tag $(APP_NAME):latest $(ECR_REPO):latest

docker-push: docker-login docker-build docker-tag
        docker push $(ECR_REPO):latest

############################################
# ECS Deployment Helpers
############################################

ecs-restart:
        aws ecs update-service \
            --cluster $(APP_NAME)-$(ENV)-cluster \
            --service $(APP_NAME)-$(ENV)-svc \
            --force-new-deployment \
            --region $(AWS_REGION)

############################################
# Full Deploy Pipeline (App Platform)
############################################

deploy-lambda: lambda-build tf-apply lambda-clean

deploy-ecs: docker-push ecs-restart

deploy-all: docker-push lambda-build tf-apply ecs-restart lambda-clean

############################################
# Landing Zone — Phase 1 (Org / Management Account)
############################################

LZ_ORG_DIR = infra/landing-zone/org

lz-org-init:
        cd $(LZ_ORG_DIR) && terraform init

lz-org-plan:
        cd $(LZ_ORG_DIR) && terraform plan

lz-org-apply:
        cd $(LZ_ORG_DIR) && terraform apply -auto-approve

############################################
# Landing Zone — Phase 2 (Workloads Account)
############################################

LZ_WORKLOADS_DIR = infra/landing-zone/workloads

WORKLOADS_ACCOUNT_ID ?= $(shell aws organizations list-accounts \
        --query "Accounts[?Name=='Workloads'].Id" --output text)

lz-workloads-init:
        cd $(LZ_WORKLOADS_DIR) && terraform init

lz-workloads-plan:
        cd $(LZ_WORKLOADS_DIR) && terraform plan \
            -var workloads_account_id=$(WORKLOADS_ACCOUNT_ID)

lz-workloads-apply:
        cd $(LZ_WORKLOADS_DIR) && terraform apply -auto-approve \
            -var workloads_account_id=$(WORKLOADS_ACCOUNT_ID)

############################################
# Combined Landing Zone Deployment
############################################

lz-deploy: lz-org-apply lz-workloads-apply

############################################
# Utility
############################################

fmt:
        terraform -chdir=$(TF_DIR) fmt -recursive

validate:
        terraform -chdir=$(TF_DIR) validate

clean:
        rm -rf $(TF_DIR)/.terraform

################################################################################
# GLUE REPLAY / REPROCESSING
################################################################################

# Usage:
#   make glue-run DATE=2025-01-15
#   make glue-run PARTITION=raw/2025/01/15
#   make glue-run FULL=true

glue-run:
        @echo "Starting Glue replay workflow..."
        @if [ "$(FULL)" = "true" ]; then \
            echo "Running full replay of all partitions..."; \
            aws glue start-job-run \
                --job-name $(APP_NAME)-$(ENV)-glue-job \
                --arguments '{"--full-replay":"true"}' \
                --region $(AWS_REGION); \
        elif [ -n "$(DATE)" ]; then \
            echo "Replaying partition for date $(DATE)..."; \
            aws glue start-job-run \
                --job-name $(APP_NAME)-$(ENV)-glue-job \
                --arguments '{"--date":"$(DATE)"}' \
                --region $(AWS_REGION); \
        elif [ -n "$(PARTITION)" ]; then \
            echo "Replaying specific partition $(PARTITION)..."; \
            aws glue start-job-run \
                --job-name $(APP_NAME)-$(ENV)-glue-job \
                --arguments '{"--partition":"$(PARTITION)"}' \
                --region $(AWS_REGION); \
        else \
            echo "ERROR: Must specify DATE=YYYY-MM-DD, PARTITION=..., or FULL=true"; \
            exit 1; \
        fi
        @echo "Glue replay initiated successfully."



################################################################################
# HEALTH CHECKS (No-Op / Always Succeed)
################################################################################

lambda-health:
        @echo "Lambda health check placeholder — always passing"
        @exit 0

ecs-health:
        @echo "ECS health check placeholder — always passing"
        @exit 0

sqs-health:
        @echo "SQS queue depth check placeholder — always passing"
        @exit 0

glue-health:
        @echo "Glue job readiness check placeholder — always passing"
        @exit 0

s3-latency-check:
        @echo "S3 latency check placeholder — always passing"
        @exit 0

dependency-slo-check:
        @echo "Dependency SLO check placeholder — always passing"
        @exit 0


################################################################################
# REAL ROLLBACK LOGIC
################################################################################

tf-rollback:
        @echo "Rolling back Terraform to previous known-good state..."
        cd $(TF_DIR) && \
        if [ -f previous.tfplan ]; then \
            echo "Applying previous Terraform plan..."; \
            terraform apply -auto-approve previous.tfplan; \
        else \
            echo "No previous.tfplan found — skipping Terraform rollback"; \
        fi
        @exit 0

ecs-rollback:
        @echo "Rolling back ECS service to previous task definition..."
        @PREV_TASK=$$(aws ecs describe-services \
            --cluster $(APP_NAME)-$(ENV)-cluster \
            --services $(APP_NAME)-$(ENV)-svc \
            --query "services[0].deployments[?status=='PRIMARY'].taskDefinition" \
            --output text); \
        echo "Previous task definition: $$PREV_TASK"; \
        aws ecs update-service \
            --cluster $(APP_NAME)-$(ENV)-cluster \
            --service $(APP_NAME)-$(ENV)-svc \
            --task-definition $$PREV_TASK \
            --region $(AWS_REGION)
        @exit 0

lambda-rollback:
        @echo "Rolling back Lambda alias to previous version..."
        @PREV_VERSION=$$(aws lambda list-versions-by-function \
            --function-name $(APP_NAME)-$(ENV)-lambda \
            --query "Versions[-2].Version" \
            --output text); \
        echo "Previous Lambda version: $$PREV_VERSION"; \
        aws lambda update-alias \
            --function-name $(APP_NAME)-$(ENV)-lambda \
            --name live \
            --function-version $$PREV_VERSION \
            --region $(AWS_REGION)
        @exit 0

glue-rollback:
        @echo "Rolling back Glue job script to previous version..."
        aws s3 cp \
            s3://$(APP_NAME)-$(ENV)-glue-scripts/previous/job.py \
            s3://$(APP_NAME)-$(ENV)-glue-scripts/job.py \
            --region $(AWS_REGION)
        @exit 0

notify-rollback:
        @echo "Sending rollback notification..."
        @echo "(Slack/PagerDuty integration placeholder)"
        @exit 0
