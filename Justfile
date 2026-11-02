dagger_version := "v0.11.9"
# Default tool versions. These fit the experiments cluster (Kubernetes 1.28.8).
# Per-cluster versions are passed as named arguments; the registry in
# cluster-ops is the source of truth (see README).
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

setup: install-gha-binary install-dagger-binary prepare-dagger-engine

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

# Prepare and verify the Dagger engine container on Linux runners
[group('setup-tools')]
prepare-dagger-engine:
    #!/usr/bin/env bash
    set -euxo pipefail
    if [ "$(uname -s)" != "Linux" ]; then
        exit 0
    fi
    old_engines=$(docker ps -aq --filter "name=^/dagger-engine-" || true)
    if [ -n "$old_engines" ]; then
        docker rm -fv $old_engines || true
    fi
    sudo modprobe ip_tables iptable_nat iptable_filter iptable_mangle ip6_tables ip6table_nat ip6table_filter br_netfilter 2>/dev/null || true
    sudo sysctl -w kernel.apparmor_restrict_unpriv_userns=0 2>/dev/null || true
    engine_image="registry.dagger.io/engine:{{dagger_version}}"
    docker pull "$engine_image"
    repo_digest=$(docker inspect --format '{{ "{{" }}index .RepoDigests 0{{ "}}" }}' "$engine_image")
    sha_hex="${repo_digest#*sha256:}"
    engine_name="dagger-engine-${sha_hex:0:16}"
    if ! grep -qx 'nat' /proc/net/ip_tables_names 2>/dev/null || ! grep -qx 'filter' /proc/net/ip_tables_names 2>/dev/null; then
        docker build -t "$engine_image" - <<EOF
    FROM ${engine_image}
    RUN for b in iptables iptables-save iptables-restore ip6tables ip6tables-save ip6tables-restore; do \
            ln -sf /sbin/xtables-nft-multi /usr/sbin/\$b && \
            ln -sf /sbin/xtables-nft-multi /sbin/\$b; \
        done
    EOF
    fi
    docker run --name "$engine_name" -d --restart always -v /var/lib/dagger --privileged "$engine_image" --debug
    ready=0
    for _ in $(seq 1 20); do
        if docker exec "$engine_name" buildctl debug workers >/dev/null 2>&1; then
            ready=1
            break
        fi
        sleep 1
    done
    if [ "$ready" -ne 1 ]; then
        docker logs "$engine_name" || true
        docker rm -fv "$engine_name" || true
        docker build -t "$engine_image" - <<EOF
    FROM ${engine_image}
    RUN for b in iptables iptables-save iptables-restore ip6tables ip6tables-save ip6tables-restore; do \
            ln -sf /sbin/xtables-nft-multi /usr/sbin/\$b && \
            ln -sf /sbin/xtables-nft-multi /sbin/\$b; \
        done
    EOF
        docker run --name "$engine_name" -d --restart always -v /var/lib/dagger --privileged "$engine_image" --debug
        for _ in $(seq 1 20); do
            if docker exec "$engine_name" buildctl debug workers >/dev/null 2>&1; then
                ready=1
                break
            fi
            sleep 1
        done
    fi
    if [ "$ready" -ne 1 ]; then
        docker ps -a --filter "name=^/dagger-engine-" || true
        docker logs "$engine_name" || true
        exit 1
    fi

# Export kubectl configuration
#
# This recipe generates a kubectl configuration file for accessing a Kubernetes
# cluster.
#
# Options:
#   --cluster                 The name of the Kubernetes cluster (required)
#   --cluster-state           The GCS bucket containing the cluster state (required)
#   --gcp-credentials-file    Path to the GCP credentials file (required)
#   --kops-version            kOps version, must match the cluster's Kubernetes
#                             minor release (default: 1.29.2)
#   --kubectl-version         kubectl version (default: 1.28.8)

[arg('cluster', long), arg('cluster-state', long), arg('gcp-credentials-file', long)]
[arg('kops-version', long), arg('kubectl-version', long)]
export-kubectl cluster cluster-state gcp-credentials-file kops-version=kops_version kubectl-version=kubectl_version: setup
    #!/usr/bin/env bash
    set -euxo pipefail
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops-version}} with-kubectl --version={{kubectl-version}} \
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
# Options:
#   --cluster                 The name of the Kubernetes cluster (required)
#   --cluster-state           The GCS bucket containing the cluster state (required)
#   --pulumi-state            The Pulumi state backend URL (required)
#   --gcp-credentials-file    Path to the GCP credentials file (required)
#   --ref                     The Git reference to deploy (required)
#   --token                   GitHub token for creating deployments (required)
#   --user                    Docker registry username (required)
#   --pass                    Docker registry password (required)
#   --kops-version            kOps version, must match the cluster's Kubernetes
#                             minor release (default: 1.29.2)
#   --kubectl-version         kubectl version (default: 1.28.8)
#   --pulumi-version          Pulumi CLI version (default: 3.108.0)
#   --cluster-ops-ref         Git ref of cluster-ops to clone (default: develop)
#   --app                     Application name (required)
#   --docker-image            Docker image name (required)
#   --docker-namespace        Docker namespace (required)
#   --dockerfile              Path to Dockerfile (required)
#   --project                 Pulumi project name (required)
#   --stack                   Pulumi stack name (required)
#   --environment             Deployment environment (required)
#   --repository              GitHub repository in owner/repo format (required)

[arg('cluster', long), arg('cluster-state', long), arg('pulumi-state', long)]
[arg('gcp-credentials-file', long), arg('ref', long), arg('token', long)]
[arg('user', long), arg('pass', long), arg('kops-version', long)]
[arg('kubectl-version', long), arg('pulumi-version', long), arg('cluster-ops-ref', long)]
[arg('app', long), arg('docker-image', long), arg('docker-namespace', long)]
[arg('dockerfile', long), arg('project', long), arg('stack', long)]
[arg('environment', long), arg('repository', long)]
deploy-buildless-backend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass kops-version=kops_version kubectl-version=kubectl_version pulumi-version=pulumi_version cluster-ops-ref="develop" app docker-image docker-namespace dockerfile project stack environment repository: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application={{app}} \
        with-docker-image --docker-image={{docker-image}} \
        with-docker-namespace --docker-namespace={{docker-namespace}} \
        with-dockerfile --dockerfile={{dockerfile}} \
        with-project --project={{project}} \
        with-stack --stack={{stack}} \
        with-environment --environment={{environment}} \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository={{repository}} \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`

    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops-version}} with-kubectl --version={{kubectl-version}} \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository={{repository}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi-version}} \
    deploy-backend-through-github --token={{token}} \
    --deployment-id=$deployment_id \
    --cluster-ops-ref={{cluster-ops-ref}}

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Deploy a backend application
#
# This recipe builds and deploys a backend application.
# It creates a GitHub deployment, builds and publishes a Docker image,
# and deploys the application using Pulumi.
#
# Options:
#   --cluster                 The name of the Kubernetes cluster (required)
#   --cluster-state           The GCS bucket containing the cluster state (required)
#   --pulumi-state            The Pulumi state backend URL (required)
#   --gcp-credentials-file    Path to the GCP credentials file (required)
#   --ref                     The Git reference to deploy (required)
#   --token                   GitHub token for creating deployments (required)
#   --user                    Docker registry username (required)
#   --pass                    Docker registry password (required)
#   --kops-version            kOps version, must match the cluster's Kubernetes
#                             minor release (default: 1.29.2)
#   --kubectl-version         kubectl version (default: 1.28.8)
#   --pulumi-version          Pulumi CLI version (default: 3.108.0)
#   --cluster-ops-ref         Git ref of cluster-ops to clone (default: develop)
#   --app                     Application name (required)
#   --docker-image            Docker image name (required)
#   --docker-namespace        Docker namespace (required)
#   --dockerfile              Path to Dockerfile (required)
#   --project                 Pulumi project name (required)
#   --stack                   Pulumi stack name (required)
#   --environment             Deployment environment (required)
#   --repository              GitHub repository in owner/repo format (required)

[arg('cluster', long), arg('cluster-state', long), arg('pulumi-state', long)]
[arg('gcp-credentials-file', long), arg('ref', long), arg('token', long)]
[arg('user', long), arg('pass', long), arg('kops-version', long)]
[arg('kubectl-version', long), arg('pulumi-version', long), arg('cluster-ops-ref', long)]
[arg('app', long), arg('docker-image', long), arg('docker-namespace', long)]
[arg('dockerfile', long), arg('project', long), arg('stack', long)]
[arg('environment', long), arg('repository', long)]
deploy-backend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass kops-version=kops_version kubectl-version=kubectl_version pulumi-version=pulumi_version cluster-ops-ref="develop" app docker-image docker-namespace dockerfile project stack environment repository: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application={{app}} \
        with-docker-image --docker-image={{docker-image}} \
        with-docker-namespace --docker-namespace={{docker-namespace}} \
        with-dockerfile --dockerfile={{dockerfile}} \
        with-project --project={{project}} \
        with-stack --stack={{stack}} \
        with-environment --environment={{environment}} \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository={{repository}} \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`

    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops-version}} with-kubectl --version={{kubectl-version}} \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    # create and publish docker image
    {{dagger_bin}} call -m {{container_module}} \
    with-repository --repository={{repository}} --should-prepend=false \
    publish-from-repo-with-deployment-id --token={{token}} \
    --user={{user}} --password={{pass}} \
    --deployment-id=$deployment_id

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository={{repository}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi-version}} \
    deploy-backend-through-github --token={{token}} \
    --deployment-id=$deployment_id \
    --cluster-ops-ref={{cluster-ops-ref}}

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Deploy a frontend application
#
# This recipe builds and deploys a frontend application.
# It creates a GitHub deployment, builds and publishes a Docker image,
# and deploys the application using Pulumi.
#
# Options:
#   --cluster                 The name of the Kubernetes cluster (required)
#   --cluster-state           The GCS bucket containing the cluster state (required)
#   --pulumi-state            The Pulumi state backend URL (required)
#   --gcp-credentials-file    Path to the GCP credentials file (required)
#   --ref                     The Git reference to deploy (required)
#   --token                   GitHub token for creating deployments (required)
#   --user                    Docker registry username (required)
#   --pass                    Docker registry password (required)
#   --kops-version            kOps version, must match the cluster's Kubernetes
#                             minor release (default: 1.29.2)
#   --kubectl-version         kubectl version (default: 1.28.8)
#   --pulumi-version          Pulumi CLI version (default: 3.108.0)
#   --cluster-ops-ref         Git ref of cluster-ops to clone (default: develop)
#   --app                     Application name (required)
#   --docker-image            Docker image name (required)
#   --docker-namespace        Docker namespace (required)
#   --dockerfile              Path to Dockerfile (required)
#   --project                 Pulumi project name (required)
#   --stack                   Pulumi stack name (required)
#   --environment             Deployment environment (required)
#   --repository              GitHub repository in owner/repo format (required)

[arg('cluster', long), arg('cluster-state', long), arg('pulumi-state', long)]
[arg('gcp-credentials-file', long), arg('ref', long), arg('token', long)]
[arg('user', long), arg('pass', long), arg('kops-version', long)]
[arg('kubectl-version', long), arg('pulumi-version', long), arg('cluster-ops-ref', long)]
[arg('app', long), arg('docker-image', long), arg('docker-namespace', long)]
[arg('dockerfile', long), arg('project', long), arg('stack', long)]
[arg('environment', long), arg('repository', long)]
deploy-frontend cluster cluster-state pulumi-state gcp-credentials-file ref token user pass kops-version=kops_version kubectl-version=kubectl_version pulumi-version=pulumi_version cluster-ops-ref="develop" app docker-image docker-namespace dockerfile project stack environment repository: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    # create github deployment
    deployment_id=`{{dagger_bin}} call -m {{gh_deployment_module}} \
        with-application --application={{app}} \
        with-docker-image --docker-image={{docker-image}} \
        with-docker-namespace --docker-namespace={{docker-namespace}} \
        with-dockerfile --dockerfile={{dockerfile}} \
        with-project --project={{project}} \
        with-stack --stack={{stack}} \
        with-environment --environment={{environment}} \
        with-kubectl-file --kubectl-file={{kubectl_file}} \
        with-repository --repository={{repository}} \
        with-ref --ref={{ref}} \
        create-github-deployment --token={{token}}`

    # set deployment to in_progress
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status=in_progress

    # generate kubectl file
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops-version}} with-kubectl --version={{kubectl-version}} \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}

    # create and publish docker image
    {{dagger_bin}} call -m {{container_module}} \
    with-repository --repository={{repository}} --should-prepend=false \
    publish-frontend-from-repo-with-deployment-id --token={{token}} \
    --user={{user}} --password={{pass}} \
    --deployment-id=$deployment_id

    #deploy the application
    {{dagger_bin}} call -m {{deploy_module}} \
    with-repository --repository={{repository}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-kube-config --config={{kubectl_file}} \
    with-backend --backend={{pulumi-state}} \
    with-pulumi --version={{pulumi-version}} \
    deploy-frontend-through-github --token={{token}} \
    --deployment-id=$deployment_id \
    --cluster-ops-ref={{cluster-ops-ref}}

    # finish with successful deployment
    {{dagger_bin}} call -m {{gh_deployment_module}} \
    with-repository --repository={{repository}} \
    set-deployment-status --token={{token}} \
    --deployment-id=$deployment_id \
    --status="success"

# Build and publish a Docker image
#
# This recipe builds and publishes a Docker image from a repository.
#
# Options:
#   --repository    GitHub repository in owner/repo format (required)
#   --ref           The Git reference to use (required)
#   --user          Docker registry username (required)
#   --pass          Docker registry password (required)
#   --namespace     Docker namespace (required)
#   --image         Docker image name (required)
#   --dockerfile    Path to Dockerfile (required)

[arg('repository', long), arg('ref', long), arg('user', long)]
[arg('pass', long), arg('namespace', long), arg('image', long)]
[arg('dockerfile', long)]
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
# Options:
#   --ref         The Git reference to use (required)
#   --user        Docker registry username (required)
#   --pass        Docker registry password (required)
#   --namespace   Docker namespace (required)
#   --image       Docker image name (required)

[arg('ref', long), arg('user', long), arg('pass', long)]
[arg('namespace', long), arg('image', long)]
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
# Options:
#   --repository    GitHub repository in owner/repo format (required)
#   --ref           The Git reference to lint (required)
#   --version       Version of the linter to use (required)

[arg('repository', long), arg('ref', long), arg('version', long)]
lint-repo repository ref version: setup
    #!/usr/bin/env bash
    set -euxo pipefail
    {{dagger_bin}} call -m golang \
        lint-git-hub \
        --repository={{repository}} \
        --git-ref={{ref}} \
        --version={{version}}
