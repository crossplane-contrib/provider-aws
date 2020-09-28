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
    pods=$("${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get pods)
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

HOSTARCH="${HOSTARCH:-amd64}"
BUILD_IMAGE="${BUILD_REGISTRY}/${PROJECT_NAME}-${HOSTARCH}"
CONTROLLER_IMAGE="${BUILD_REGISTRY}/${PROJECT_NAME}-controller-${HOSTARCH}"

version_tag="$(cat ${projectdir}/_output/version)"
# tag as latest version to load into kind cluster
PACKAGE_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}:${VERSION}"
PACKAGE_CONTROLLER_IMAGE="${DOCKER_REGISTRY}/${PROJECT_NAME}-controller:${VERSION}"
K8S_CLUSTER="${K8S_CLUSTER:-${BUILD_REGISTRY}-INTTESTS}"

CROSSPLANE_NAMESPACE="crossplane-system"
PACKAGE_NAME="provider-aws"

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

# tag package and controller images and load then into kind cluster
docker tag "${BUILD_IMAGE}" "${PACKAGE_IMAGE}"
docker tag "${CONTROLLER_IMAGE}" "${PACKAGE_CONTROLLER_IMAGE}"
"${KIND}" load docker-image "${PACKAGE_IMAGE}" --name="${K8S_CLUSTER}"
"${KIND}" load docker-image "${PACKAGE_CONTROLLER_IMAGE}" --name="${K8S_CLUSTER}"

# wait for kind pods
echo_step "wait for kind pods"
kindpods=("kube-apiserver-${BUILD_REGISTRY}-inttests-control-plane" "kube-controller-manager-${BUILD_REGISTRY}-inttests-control-plane" "kube-scheduler-${BUILD_REGISTRY}-inttests-control-plane")
wait_for_pods_in_namespace 120 "kube-system" "${kindpods[@]}"

# install crossplane from master channel
echo_step "installing crossplane from master channel"
"${KUBECTL}" create ns crossplane-system
"${HELM3}" repo add crossplane-master https://charts.crossplane.io/master/
chart_version="$("${HELM3}" search repo crossplane-master/crossplane --devel | awk 'FNR == 2 {print $2}')"
echo_info "using crossplane version ${chart_version}"
echo
"${HELM3}" install crossplane --namespace crossplane-system crossplane-master/crossplane --version ${chart_version} --devel

echo_step "waiting for deployment crossplane rollout to finish"
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" rollout status "deploy/crossplane" --timeout=2m

echo_step "wait until the pods are up and running"
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" wait --for=condition=Ready pods --all --timeout=1m

# ----------- integration tests
echo_step "--- INTEGRATION TESTS ---"
echo
echo_step "check for necessary deployment statuses"
echo
echo -------- deployments
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get deployments

check_deployments "crossplane" "${CROSSPLANE_NAMESPACE}"

echo_step "check for crossplane pods statuses"
echo
echo "--- pods ---"
check_pods $("${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get pods)

# install package
echo_step "installing ${PROJECT_NAME} into \"${CROSSPLANE_NAMESPACE}\" namespace"

INSTALL_YAML="$( cat <<EOF
apiVersion: pkg.crossplane.io/v1alpha1
kind: Provider
metadata:
  name: "${PACKAGE_NAME}"
spec:
  package: "${PACKAGE_IMAGE}"
  packagePullPolicy: Never
EOF
)"

echo "${INSTALL_YAML}" | "${KUBECTL}" apply -f -

"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get deployments
"${KUBECTL}" -n "${CROSSPLANE_NAMESPACE}" get pods

# this is to let package manager unpack pods and start the controller. If for
# some reason the package manager did not create these pods at all, then this
# could give a false positive.
sleep 20

echo_step "check for package pod statuses"
echo
echo "--- pods ---"
check_pods

echo_step "uninstalling ${PROJECT_NAME}"

echo "${INSTALL_YAML}" | "${KUBECTL}" delete -f -

# check pods deleted

echo_success "Integration tests succeeded!"
