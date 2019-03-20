# Template to deploy Aqua Enforcer

# All option in one Template (Don't use it please delete the options that you don't use)

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaEnforcer
metadata:
  name: aqua
  namespace: aqua
spec:
  requirements: false
  serviceAccount: "aqua-sa" # only if requirments false
  registry: # only if requirements true
    url: "registry.aquasec.com"
    username: "example@company.com"
    password: "" # Password
    email: "example@company.com"
  token: "tests"
  rbac:
    enable: false
    roleref: ""
    openshift: false
    privileged: true
  deploy:
    image:
      repository: "enforcer"
      registry: "registry.aquasec.com"
      tag: "4.0"
      pullPolicy: "IfNotPresent"
  gateway:
    host: aqua-gateway-svc
    port: 3622
  sendingHostImages: false
  runcInterception: false
```

### Template for example

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaEnforcer
metadata:
  name: aqua
spec:
  requirements: false
  serviceAccount: "aqua-sa"
  token: "tests"
  rbac:
    enable: false
    openshift: false
    privileged: true
  deploy:
    image:
      repository: "enforcer"
      registry: "registry.aquasec.com"
      tag: "4.0"
      pullPolicy: "IfNotPresent"
  gateway:
    host: aqua-gateway-svc
    port: 3622
  sendingHostImages: false
  runcInterception: false
```