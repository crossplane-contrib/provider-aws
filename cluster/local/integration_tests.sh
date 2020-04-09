#!/usr/bin/env bash
set -e

# setting up colors
BLU='\033[0;34m'
YLW='\033[0;33m'
GRN='\033[0;32m'
RED='\033[0;31m'
NOC='\033[0m' # No Color
echo_info(){
    printf "\n${BLU}%s${NOC}" "$1"
}
echo_step(){
    printf "\n${BLU}>>>>>>> %s${NOC}\n" "$1"
}
echo_sub_step(){
    printf "\n${BLU}>>> %s${NOC}\n" "$1"
}

echo_step_completed(){
    printf "${GRN} [✔]${NOC}"
}

echo_success(){
    printf "\n${GRN}%s${NOC}\n" "$1"
}
echo_warn(){
    printf "\n${YLW}%s${NOC}" "$1"
}
echo_error(){
    printf "\n${RED}%s${NOC}" "$1"
    exit 1
}

# k8s watchers
wait_for_deployment_create(){
    local timeout=$1
    local counter=0
    echo -n "waiting for deployment create in $2..." >&2
    while "${KUBECTL}" -n $2 get deployments -o yaml | grep -q  'items: \[\]'; do
        if [ "$counter" -ge "$timeout" ]; then echo "TIMEOUT"; exit -1; else (( counter+=5 )); fi
        echo -n "." >&2
        sleep 5
    done
}

wait_for_deployment_delete(){
    local timeout=$1
    local counter=0
    echo -n "waiting for deployment delete in $2..." >&2
    while ! "${KUBECTL}" -n $2 get deployments -o yaml | grep -q  'items: \[\]'; do
        if [ "$counter" -ge "$timeout" ]; then echo "TIMEOUT"; exit -1; else (( counter+=5 )); fi
        echo -n "." >&2
        sleep 5
    done
}

wait_for_pods_in_namespace(){
    local timeout=$1
    shift
    namespace=$1
    shift
    arr=("$@")
    local counter=0
    for i in "${arr[@]}";
        do
            echo -n "waiting for pod $i in namespace $namespace..." >&2
            while ! ("${KUBECTL}" -n $namespace get pod $i) &>/dev/null; do
                if [ "$counter" -ge "$timeout" ]; then echo "TIMEOUT"; exit -1; else (( counter+=5 )); fi
                echo -n "." >&2
                sleep 5
            done
            echo "FOUND POD!" >&2
        done
}

check_deployments(){
    for name in $1; do
        echo_sub_step "inspecting deployment '${name}'"
        local dep_stat=$("${KUBECTL}" -n "$2" get deployments/"${name}")

        echo_info "check if is deployed"
        if $(echo "$dep_stat" | grep -iq 'No resources found'); then
            echo "is not deployed"
            exit -1
        else
            echo_step_completed
        fi

        echo_info "check if is ready"
        IFS='/' read -ra ready_status_parts <<< "$(echo "$dep_stat" | awk ' FNR > 1 {print $2}')"
        if (("${ready_status_parts[0]}" < "${ready_status_parts[1]}")); then
            echo "is not Ready"
            exit -1
        else
            echo_step_completed
        fi
        echo
    done
}

check_pods(){
    pods=$("${KUBECTL}" -n "$1" get pods)
    echo "$pods"
    while read -r pod_stat; do
        name=$(echo "$pod_stat" | awk '{print $1}')
        echo_sub_step "inspecting pod '${name}'"

        if $(echo "$pod_stat" | awk '{print $3}' | grep -ivq 'Completed'); then
            echo_info "is not completed, continuing with further checks"
        else
            echo_info "is completed, foregoing further checks"
            echo_step_completed
            continue
        fi

        echo_info "check if is ready"
        IFS='/' read -ra ready_status_parts <<< "$(echo "$pod_stat" | awk '{print $2}')"
        if (("${ready_status_parts[0]}" < "${ready_status_parts[1]}")); then
            echo_error "is not ready"
            exit -1
        else
            echo_step_completed
        fi

        echo_info "check if is running"
        if $(echo "$pod_stat" | awk '{print $3}' | grep -ivq 'Running'); then
            echo_error "is not running"
            exit -1
        else
            echo_step_completed
        fi

        echo_info "check if has restarts"
        if (( $(echo "$pod_stat" | awk '{print $4}') > 0 )); then
            echo_error "has restarts"
            exit -1
        else
            echo_step_completed
        fi
        echo
    done <<< "$(echo "$pods" | awk 'FNR>1')"
}

# ------------------------------
projectdir="$( cd "$( dirname "${BASH_SOURCE[0]}")"/../.. && pwd )"

# get the build environment variables from the special build.vars target in the main makefile
eval $(make --no-print-directory -C ${projectdir} build.vars)

# ------------------------------

HELM_VERSION="helm-v2.16.0"
HOST_PLATFORM="${HOSTOS}_${HOSTARCH}"
TOOLS_DIR="${CACHE_DIR}/tools"
TOOLS_HOST_DIR="${TOOLS_DIR}/${HOST_PLATFORM}"
HELM="${TOOLS_HOST_DIR}/${HELM_VERSION}"
echo_step "installing ${HELM_VERSION} ${HOSTOS}-${HOSTARCH}"
mkdir -p ${TOOLS_HOST_DIR}/tmp-helm
curl -fsSL https://storage.googleapis.com/kubernetes-helm/${HELM_VERSION}-${HOSTOS}-${HOSTARCH}.tar.gz | tar -xz -C ${TOOLS_HOST_DIR}/tmp-helm
mv ${TOOLS_HOST_DIR}/tmp-helm/${HOSTOS}-${HOSTARCH}/helm ${HELM}
rm -fr ${TOOLS_HOST_DIR}/tmp-helm
echo_success "installing ${HELM_VERSION} ${HOSTOS}-${HOSTARCH}"

HOSTARCH="${HOSTARCH:-amd64}"
BUILD_IMAGE="${BUILD_REGISTRY}/${PROJECT_NAME}-${HOSTARCH}"

version_tag="$(cat ${projectdir}/_output/version)"
# tag as master so that config/stack/install.yaml can be used
STACK_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}:master"
K8S_CLUSTER="${K8S_CLUSTER:-${BUILD_REGISTRY}-INTTESTS}"

CROSSPLANE_NAMESPACE="crossplane-system"
STACK_NAME="provider-aws"
STACK_NAMESPACE="aws"

# cleanup on exit
if [ "$skipcleanup" != true ]; then
  function cleanup {
    echo_step "Cleaning up..."
    export KUBECONFIG=
    "${KIND}" delete cluster --name="${K8S_CLUSTER}"
  }

  trap cleanup EXIT
fi

echo_step "creating k8s cluster using kind"
"${KIND}" create cluster --name="${K8S_CLUSTER}"

# tag stack image and load it to kind cluster
docker tag "${BUILD_IMAGE}" "${STACK_IMAGE}"
"${KIND}" load docker-image "${STACK_IMAGE}" --name="${K8S_CLUSTER}"

echo_step "installing tiller"
"${KUBECTL}" apply -f "${projectdir}/cluster/local/helm-rbac.yaml"
"${HELM}" init --service-account tiller
# waiting for deployment "tiller-deploy" rollout to finish
"${KUBECTL}" -n kube-system rollout status deploy/tiller-deploy --timeout=2m

# wait for kind pods
echo_step "wait for kind pods"
kindpods=("kube-apiserver-${BUILD_REGISTRY}-inttests-control-plane" "kube-controller-manager-${BUILD_REGISTRY}-inttests-control-plane" "kube-scheduler-${BUILD_REGISTRY}-inttests-control-plane")
wait_for_pods_in_namespace 120 "kube-system" "${kindpods[@]}"

# install crossplane from master channel
echo_step "installing crossplane from master channel"
"${HELM}" repo add crossplane-master https://charts.crossplane.io/master/
chart_version="$("${HELM}" search crossplane-master/crossplane | awk 'FNR == 2 {print $2}')"
"${HELM}" install --name crossplane --namespace crossplane-system crossplane-master/crossplane --version ${chart_version}

echo_step "waiting for deployment crossplane rollout to finish"
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" rollout status "deploy/crossplane" --timeout=2m

echo_step "wait until the pods are up and running"
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" wait --for=condition=Ready pods --all --timeout=1m

# ----------- integration tests
echo_step "------------------------------ INTEGRATION TESTS"
echo
echo_step "check for necessary deployment statuses"
echo
echo -------- deployments
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get deployments

check_deployments "crossplane crossplane-stack-manager" "${CROSSPLANE_NAMESPACE}"

echo_step "check for crossplane pods statuses"
echo
echo "-------- pods"
check_pods "${CROSSPLANE_NAMESPACE}"

# let stack manager initialize controllers and workers
sleep 30

# install stack into stack namespace
echo_step "installing ${PROJECT_NAME} into \"${STACK_NAMESPACE}\" namespace"

INSTALL_YAML="$( cat <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: "${STACK_NAMESPACE}"
---
apiVersion: stacks.crossplane.io/v1alpha1
kind: ClusterStackInstall
metadata:
  name: "${STACK_NAME}"
  namespace: "${STACK_NAMESPACE}"
spec:
  package: "${STACK_IMAGE}"
EOF
)"

echo "${INSTALL_YAML}" | "${KUBECTL}" apply -f -

wait_for_deployment_create 120 "${STACK_NAMESPACE}"

"${KUBECTL}" -n "${STACK_NAMESPACE}" get deployments
"${KUBECTL}" -n "${STACK_NAMESPACE}" get pods

echo_step "waiting for deployment ${STACK_NAME} rollout to finish"
"${KUBECTL}" -n "${STACK_NAMESPACE}" rollout status "deploy/${PROJECT_NAME}-controller" --timeout=2m

check_deployments "${STACK_NAME}-controller" "${STACK_NAMESPACE}"

echo_step "check for stack pods statuses"
echo
echo "-------- pods"
check_pods "${STACK_NAMESPACE}"

echo_step "uninstalling ${PROJECT_NAME}"

echo "${INSTALL_YAML}" | "${KUBECTL}" delete -f -

wait_for_deployment_delete 120 "${STACK_NAMESPACE}"

echo_success "Integration tests succeeded!"
