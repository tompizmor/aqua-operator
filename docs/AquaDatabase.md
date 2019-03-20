# Template to deploy Aqua Database

# All option in one Template (Don't use it please delete the options that you don't use)

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaDatabase
metadata:
  name: aqua
  namespace: aqua
spec:
  requirements: true
  serviceAccount: "" # if requirements false and service account allready exists
  dbSecretName: "aqua-database-password" # if requirements false and db secret is exists
  dbSecretKey: "db-password" # if requirements false and db secret is exists
  password: "Password1" # only if requirements true DB Password Secret
  registry: # only if requirements true
    url: "registry.aquasec.com"
    username: "example@company.com"
    password: "" # Password
    email: "example@company.com"
  deploy:
    replicas: 1
    service: "ClusterIP"
    image:
      repository: "database"
      registry: "registry.aquasec.com"
      tag: "3.5"
      pullpolicy: "IfNotPresent"
  openshift: false
```

### Template for example

```yaml
apiVersion: operator.aquasec.com/v1alpha1
kind: AquaDatabase
metadata:
  name: aqua
spec:
  requirements: true
  password: "Password1"
  registry:
    url: "registry.aquasec.com"
    username: "example@gmail.com"
    password: "" # Password
    email: "example@gmail.com"
  deploy:
    replicas: 1
    service: "ClusterIP"
    image:
      repository: "database"
      registry: "registry.aquasec.com"
      tag: "3.5"
      pullpolicy: "IfNotPresent"
  openshift: false
```