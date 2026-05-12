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
