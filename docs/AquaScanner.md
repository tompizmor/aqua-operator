# Template to deploy Aqua Gateway

# All option in one Template (Don't use it please delete the options that you don't use)

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaScanner
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
  deploy:
    replicas: 1
    service: "ClusterIP"
    image:
      repository: "scanner"
      registry: "registry.aquasec.com"
      tag: "4.0"
      pullPolicy: "IfNotPresent"
  login: 
    username: "administrator"
    password: "Password1"
    host: "http://aqua-server-svc:8080"
  openshift: false
```

### Template for example

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaScanner
metadata:
  name: aqua
spec:
  requirements: false
  serviceAccount: "aqua-sa"
  deploy:
    replicas: 1
    service: "ClusterIP"
    image:
      repository: "scanner"
      registry: "registry.aquasec.com"
      tag: "4.0"
      pullPolicy: "IfNotPresent"
  login:
    username: "administrator"
    password: "Password1"
    host: "http://aqua-server-svc:8080"
  openshift: false
```