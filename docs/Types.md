# Types

- [Types](#types)
  - [Aqua Docker Registry (AquaDockerRegistry)](#aqua-docker-registry-aquadockerregistry)
  - [Aqua Database Information (AquaDatabaseInformation)](#aqua-database-information-aquadatabaseinformation)
  - [Aqua Docker Image Details (AquaImage)](#aqua-docker-image-details-aquaimage)
  - [Aqua Service (AquaService)](#aqua-service-aquaservice)
  - [Aqua Rbac Settings (AquaRbacSettings)](#aqua-rbac-settings-aquarbacsettings)
  - [Aqua Gateway Information (AquaGatewayInformation)](#aqua-gateway-information-aquagatewayinformation)
  - [Aqua Login (AquaLogin)](#aqua-login-aqualogin)
  - [Aqua Scanner CLI Scale (AquaScannerCliScale)](#aqua-scanner-cli-scale-aquascannercliscale)

## Aqua Docker Registry (AquaDockerRegistry)

Properties:
* **Url** - The docker registry URL for example: registry.aquasec.com, docker.io and etc
* **Username** - User name
* **Password** - Password
* **Email** - User's email

## Aqua Database Information (AquaDatabaseInformation)

Properties:
* **Name** - PostgreSQL Database Name
* **Host** - PostgreSQL Host name or IP Address
* **Port** - PostgreSQL Database Port (default: 5432)
* **Username** - User name
* **Password** - User's password
* **AuditName** - PostgreSQL Audit Database Name
* **AuditHost** - PostgreSQL Audit Host name or IP Address
* **AuditPort** - PostgreSQL Audit Database Port (default: 5432)
* **AuditUsername** - Audit User name
* **AuditPassword** - Audit User's password

## Aqua Docker Image Details (AquaImage)

Properties: 
* **Repository** - Docker Repository Name
* **Registry** - Docker Registry Name
* **Tag** - Image Tag
* **PullPolicy** - Docker Image Pull Policy (Same as Kubernetes)
  
## Aqua Service (AquaService)
AquaService Struct for deployment spec

Properties:
* **Replicas** - Number of Replicas
* **ServiceType** - Type of service (NodePort, ClusterIP or LoadBalancer)
* **ImageData** - Information about the docker image (Type - *AquaImage*)
* **Resources** - Resources settings
* **LivenessProbe** - LivenessProbe settings
* **ReadinessProbe** - RedinessProbe settings
* **NodeSelector** - NodeSelector settings
* **Affinity** - Affinity settings
* **Tolerations** - Tolerations settings

## Aqua Rbac Settings (AquaRbacSettings)

Properties:
* **Enable** - Enable RBAC (true/false)
* **RoleRef** - Rbac Role Reference if exists
* **Openshift** - Deployment on Openshift?! (true/false)
* **Privileged** - Use Privileged (true/false)

## Aqua Gateway Information (AquaGatewayInformation)

Properties:
* **Host** - Gateway host name or K8s Service or IP Address
* **Port** - Gateway port

## Aqua Login (AquaLogin)

Properties:
* **Username** - Username for Aqua
* **Password** - Password
* **Host** - Aqua Server URL or Host with port

## Aqua Scanner CLI Scale (AquaScannerCliScale)

Properties:
* **Deploy** - Aqua Service Details about scanner cli deployment data
* **Name** - Existing Aqua Scanner Name for Scaling
* **Max** - Max Scanners per Host
* **Min** - Min Scanners per Host 
* **ImagesPerScanner** - Images per Scanner