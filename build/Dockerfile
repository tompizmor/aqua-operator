#FROM alpine:3.8
FROM registry.access.redhat.com/devtools/go-toolset-1.10-rhel7

#RUN apk upgrade --update --no-cache
#USER nobody

LABEL io.k8s.display-name="OpenShift / Kubernetes Aqua Security operator" \
      io.k8s.description="Aqua Security Operator to deploy aqua component and maintenance" \
      maintainer="Nissim Bitan <nissim.bitan@aquasec.com>"

ADD build/_output/bin/aqua-operator /usr/local/bin/aqua-operator
