APP_NAME := hesp-dev-ingest
ECR_REGISTRY := 045797643729.dkr.ecr.us-east-1.amazonaws.com
ECR_REPO := $(ECR_REGISTRY)/$(APP_NAME)
DOCKERFILE := cmd/ingest-service/Dockerfile

# Default target
all: build

# Build local image
build:
    docker build -f $(DOCKERFILE) -t $(APP_NAME):latest .

# Tag image for ECR
tag:
    docker tag $(APP_NAME):latest $(ECR_REPO):latest

# Login to ECR
login:
    aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $(ECR_REGISTRY)

# Push image to ECR
push: tag login
    docker push $(ECR_REPO):latest

# Build + push
deploy: build push

# Force ECS service to redeploy
ecs-redeploy:
    aws ecs update-service --cluster hesp-dev-cluster --service hesp-dev-svc --force-new-deployment

# Full pipeline
release: deploy ecs-redeploy
