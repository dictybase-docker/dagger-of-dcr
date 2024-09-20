set dotenv-load
dagger_version := "v0.11.9"
pulumi_version := "3.108.0"
kops_version := "1.27.1"
kops_module := "kops"
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

system-info:
    @echo this is an {{arch()}} os {{os()}}

check-env:
	@echo $DOCKERFILE $DOCKER_IMAGE $DOCKER_NAMESPACE
	@echo $REPOSITORY $ENVIRONMENT $PROJECT $STACK $APP

setup: install-gha-binary install-dagger-binary
[group('setup-tools')]
install-gha-binary:
	@curl -L -o {{action_bin}} {{gha_download_url}}
	@chmod +x {{action_bin}} 
[group('setup-tools')]
install-dagger-binary:
	{{action_bin}} sd --dagger-version {{dagger_version}} --dagger-bin-dir {{bin_path}} --dagger-file {{dagger_file}}

export-kubectl cluster cluster-state gcp-credentials-file: setup
    #!/usr/bin/env bash
    set -euxo pipefail
    {{dagger_bin}} call -m {{kops_module}} \
    with-kops --version={{kops_version}} with-kubectl \
    with-state-storage --storage={{cluster-state}} \
    with-credentials --credentials={{gcp-credentials-file}} \
    with-cluster --name={{cluster}} \
    export-kubectl --output={{kubectl_file}}


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
    with-repository --repository=$REPOSITORY \
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
    with-repository --repository=$REPOSITORY \
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

build-publish-arangopg-image ref user pass namespace image: setup
    #!/usr/bin/env bash
    set -euxo pipefail

    {{dagger_bin}} call -m {{container_module}} \
    with-ref --ref={{ref}} \
    with-namespace --namespace={{namespace}} \
    with-image --image={{image}} \
    build-and-publish-arango-postgres-container \
    --user={{user}} --password={{pass}}
