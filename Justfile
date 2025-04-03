set dotenv-load
dagger_version := "v0.11.9"
pulumi_version := "3.108.0"
kops_version := "1.29.2"
kops_module := "kops"
kubectl_version := "1.28.8"
gh_deployment_module := "gh-deployment"
container_module := "container-image"
deploy_module := "pulumi-ops"
bin_path := `mktemp -d`
action_bin := bin_path + "/actions"
dagger_bin := bin_path + "/dagger"
kubectl_file := `mktemp -d` + "/dictycr.yaml"

base_gha_download_url := "https://github.com/dictybase-docker/github-actions/releases/download/v2.10.0/action_2.10.0_"
gha_download_url := if os() == "macos" {
    base_gha_download_url + "darwin_arm64"
} else {
    base_gha_download_url + "linux_amd64"
}

file_suffix := ".tar.gz"
dagger_file := if os() == "macos" {
    "darwin_arm64" + file_suffix
} else {
    "linux_amd64" + file_suffix
}

# Display system information
#
# This recipe prints the architecture and operating system of the current machine.
system-info:
    @echo this is an {{arch()}} os {{os()}}

# Check environment variables
#
# This recipe displays the values of important environment variables
# used by other recipes for deployment and configuration.
check-env:
	@echo $DOCKERFILE $DOCKER_IMAGE $DOCKER_NAMESPACE
	@echo $REPOSITORY $ENVIRONMENT $PROJECT $STACK $APP

# Set up required tools
#
# This recipe installs the necessary binaries for GitHub Actions and Dagger.

setup: install-gha-binary install-dagger-binary

# Install GitHub Actions binary
#
# Downloads and installs the GitHub Actions binary appropriate for the current
# platform.

[group('setup-tools')]
install-gha-binary:
	@curl -L -o {{action_bin}} {{gha_download_url}}
	@chmod +x {{action_bin}} 

# Install Dagger binary
#
# Downloads and installs the Dagger binary appropriate for the current platform.

[group('setup-tools')]
install-dagger-binary:
	{{action_bin}} sd --dagger-version {{dagger_version}} --dagger-bin-dir {{bin_path}} --dagger-file {{dagger_file}}

# Export kubectl configuration
#
# This recipe generates a kubectl configuration file for accessing a Kubernetes
# cluster.
#
# Args:
#   cluster: The name of the Kubernetes cluster
#   cluster-state: The GCS bucket containing the cluster state
#   gcp-credentials-file: Path to the GCP credentials file

export-kubectl cluster cluster-state gcp-credentials-file: setup
    #!/usr/bin/env bash
    set -euxo pipefail
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops_version}} with-kubectl --version={{kubectl_version}} \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}


# Deploy a backend application without building a new Docker image
# 
# This recipe deploys a backend application using an existing Docker image.
# It creates a GitHub deployment, sets up Kubernetes configuration, and deploys
# the application using Pulumi without building a new Docker image.
#
# Args:
#   cluster: The name of the Kubernetes cluster
#   cluster-state: The GCS bucket containing the cluster state
#   pulumi-state: The Pulumi state backend URL
#   gcp-credentials-file: Path to the GCP credentials file
#   ref: The Git reference to deploy
#   token: GitHub token for creating deployments
#   user: Docker registry username
#   pass: Docker registry password
#
# Environment variables required:
#   APP: Application name
#   DOCKER_IMAGE: Docker image name
#   DOCKER_NAMESPACE: Docker namespace
#   DOCKERFILE: Path to Dockerfile
#   PROJECT: Pulumi project name
#   STACK: Pulumi stack name
#   ENVIRONMENT: Deployment environment
#   REPOSITORY: GitHub repository in owner/repo format

deploy-buildless-backend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application=$APP \
        with-docker-image --docker-image=$DOCKER_IMAGE \
        with-docker-namespace --docker-namespace=$DOCKER_NAMESPACE \
        with-dockerfile --dockerfile=$DOCKERFILE \
        with-project --project=$PROJECT \
        with-stack --stack=$STACK \
        with-environment --environment=$ENVIRONMENT \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository=$REPOSITORY \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`
    
    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops_version}} with-kubectl \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository=$REPOSITORY \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi_version}} \
    deploy-backend-through-github --token={{token}} \
    --deployment-id=$deployment_id

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Deploy a backend application
#
# This recipe builds and deploys a backend application.
# It creates a GitHub deployment, builds and publishes a Docker image,
# and deploys the application using Pulumi.
#
# Args:
#   cluster: The name of the Kubernetes cluster
#   cluster-state: The GCS bucket containing the cluster state
#   pulumi-state: The Pulumi state backend URL
#   gcp-credentials-file: Path to the GCP credentials file
#   ref: The Git reference to deploy
#   token: GitHub token for creating deployments
#   user: Docker registry username
#   pass: Docker registry password
#
# Environment variables required:
#   APP: Application name
#   DOCKER_IMAGE: Docker image name
#   DOCKER_NAMESPACE: Docker namespace
#   DOCKERFILE: Path to Dockerfile
#   PROJECT: Pulumi project name
#   STACK: Pulumi stack name
#   ENVIRONMENT: Deployment environment
#   REPOSITORY: GitHub repository in owner/repo format

deploy-backend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application=$APP \
        with-docker-image --docker-image=$DOCKER_IMAGE \
        with-docker-namespace --docker-namespace=$DOCKER_NAMESPACE \
        with-dockerfile --dockerfile=$DOCKERFILE \
        with-project --project=$PROJECT \
        with-stack --stack=$STACK \
        with-environment --environment=$ENVIRONMENT \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository=$REPOSITORY \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`
    
    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops_version}} with-kubectl \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    # create and publish docker image
    {{dagger_bin}} call -m {{container_module}} \
    with-repository --repository=$REPOSITORY --should-prepend=false \
    publish-from-repo-with-deployment-id --token={{token}} \
    --user={{user}} --password={{pass}} \
    --deployment-id=$deployment_id

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository=$REPOSITORY \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi_version}} \
    deploy-backend-through-github --token={{token}} \
    --deployment-id=$deployment_id

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Deploy a frontend application
#
# This recipe builds and deploys a frontend application.
# It creates a GitHub deployment, builds and publishes a Docker image,
# and deploys the application using Pulumi.
#
# Args:
#   cluster: The name of the Kubernetes cluster
#   cluster-state: The GCS bucket containing the cluster state
#   pulumi-state: The Pulumi state backend URL
#   gcp-credentials-file: Path to the GCP credentials file
#   ref: The Git reference to deploy
#   token: GitHub token for creating deployments
#   user: Docker registry username
#   pass: Docker registry password
#
# Environment variables required:
#   APP: Application name
#   DOCKER_IMAGE: Docker image name
#   DOCKER_NAMESPACE: Docker namespace
#   DOCKERFILE: Path to Dockerfile
#   PROJECT: Pulumi project name
#   STACK: Pulumi stack name
#   ENVIRONMENT: Deployment environment
#   REPOSITORY: GitHub repository in owner/repo format

deploy-frontend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application=$APP \
        with-docker-image --docker-image=$DOCKER_IMAGE \
        with-docker-namespace --docker-namespace=$DOCKER_NAMESPACE \
        with-dockerfile --dockerfile=$DOCKERFILE \
        with-project --project=$PROJECT \
        with-stack --stack=$STACK \
        with-environment --environment=$ENVIRONMENT \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository=$REPOSITORY \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`
    
    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops_version}} with-kubectl \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    # create and publish docker image
    {{dagger_bin}} call -m {{container_module}} \
    with-repository --repository=$REPOSITORY --should-prepend=false \
    publish-frontend-from-repo-with-deployment-id --token={{token}} \
    --user={{user}} --password={{pass}} \
    --deployment-id=$deployment_id

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository=$REPOSITORY \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi_version}} \
    deploy-frontend-through-github --token={{token}} \
    --deployment-id=$deployment_id

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository=$REPOSITORY \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Build and publish a Docker image
#
# This recipe builds and publishes a Docker image from a repository.
#
# Args:
#   repository: GitHub repository in owner/repo format
#   ref: The Git reference to use
#   user: Docker registry username
#   pass: Docker registry password
#   namespace: Docker namespace
#   image: Docker image name
#   dockerfile: Path to Dockerfile

build-publish-image repository ref user pass namespace image dockerfile: setup 
    #!/usr/bin/env bash
    set -euxo pipefail

    {{dagger_bin}} call -m {{container_module}} \
    with-ref --ref={{ref}} \
    with-namespace --namespace={{namespace}} \
    with-image --image={{image}} \
    with-dockerfile --docker-file={{dockerfile}} \
    with-repository --repository={{repository}} \
    publish-from-repo \
    --user={{user}} --password={{pass}} 

# Build and publish ArangoDB with PostgreSQL image
#
# This recipe builds and publishes a specialized Docker image containing
# both ArangoDB and PostgreSQL.
#
# Args:
#   ref: The Git reference to use
#   user: Docker registry username
#   pass: Docker registry password
#   namespace: Docker namespace
#   image: Docker image name

build-publish-arangopg-image ref user pass namespace image: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    {{dagger_bin}} call -m {{container_module}} \
    with-ref --ref={{ref}} \
    with-namespace --namespace={{namespace}} \
    with-image --image={{image}} \
    build-and-publish-arango-postgres-container \
    --user={{user}} --password={{pass}}

# Lint a repository
#
# This recipe runs linting checks on a GitHub repository.
#
# Args:
#   repository: GitHub repository in owner/repo format
#   ref: The Git reference to lint

lint-repo repository ref version: setup
    #!/usr/bin/env bash
    set -euxo pipefail
    {{dagger_bin}} call -m golang \
        lint-git-hub \
        --repository={{repository}} \
        --git-ref={{ref}} \
        --version={{version}}

