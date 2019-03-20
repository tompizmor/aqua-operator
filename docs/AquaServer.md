# Template to deploy Aqua Server

# All option in one Template (Don't use it please delete the options that you don't use)

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaServer
metadata:
  name: aqua
  namespace: aqua
spec:
  requirements: false
  add: false # only if allready exists aqua server and you want to add server to cluster
  serviceAccount: "aqua-sa" # only if requirements false and service account exists
  dbSecretName: "aqua-database-password" # only if requirements false and secret exists
  dbSecretKey: "db-password" # only if requirements false and secret exists
  password: "Password1" # only if requirements true
  registry: # only if requirements true
    url: "registry.aquasec.com"
    username: "example@company.com"
    password: "" # Password
    email: "example@company.com"
  rbac:
    enable: false
    roleref: ""
    openshift: false
    privileged: true
  deploy:
    replicas: 1
    service: "LoadBalancer"
    image:
      repository: "console"
      registry: "registry.aquasec.com"
      tag: "3.5"
      pullPolicy: "IfNotPresent"
  aquaDb: "aqua-database" # if aqua database allready exists
  externalDb: # If allready exists postgresql db and want to use an external DB
    name: "scalock"
    host: "aqua-database-svc"
    port: 5432
    username: "postgres"
    password: "example"
    auditName: "slk_audit"
    auditHost: "aqua-database-svc"
    auditPort: 5432
    auditUsername: "postgres"
    auditPassword: "example"
  adminPassword: "Password1"
  licenseToken: # Licence
  clusterMode: true
  ssl: false
  auditSsl: false
  sslCertPath: ""
  dockerless: false
  openshift: false
```

### Template for example

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaServer
metadata:
  name: aqua
spec:
  requirements: false
  add: false
  serviceAccount: "aqua-sa"
  dbSecretName: "aqua-database-password"
  dbSecretKey: "db-password"
  rbac:
    enable: true
    openshift: false
    privileged: true
  deploy:
    replicas: 1
    service: "LoadBalancer"
    image:
      repository: "console"
      registry: "registry.aquasec.com"
      tag: "3.5"
      pullPolicy: "IfNotPresent"
  aquaDb: "aqua-database" 
  adminPassword: "Password1"
  licenseToken: # Licence
  clusterMode: true
  ssl: false
  auditSsl: false
  dockerless: false
  openshift: false   
```