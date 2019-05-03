FROM registry.access.redhat.com/devtools/go-toolset-1.10-rhel7

LABEL name="Aqua Operator" \
      vendor="Aqua Security Software Ltd." \
      version="v0.0.1" \
      release="1" \
      summary="Aqua Security Operator." \
      description="The Aqua Security Operator runs within a Openshift (or Kubernetes) cluster, and provides a means to deploy and manage the Aqua Security cluster and components"

COPY licenses /licenses

ADD build/_output/bin/aqua-operator /usr/local/bin/aqua-operator
