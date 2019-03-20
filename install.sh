kubectl create -f  deploy/crds/operator_v1alpha1_aquadatabase_crd.yaml 
kubectl create -f  deploy/crds/operator_v1alpha1_aquagateway_crd.yaml 
kubectl create -f  deploy/crds/operator_v1alpha1_aquaserver_crd.yaml 
kubectl create -f  deploy/crds/operator_v1alpha1_aquaenforcer_crd.yaml
kubectl create -f  deploy/crds/operator_v1alpha1_aquacsp_crd.yaml 
kubectl create -f  deploy/crds/operator_v1alpha1_aquascanner_crd.yaml

kubectl create secret docker-registry aqua-registry-secret --docker-server=registry.aquasec.com --docker-username=<AQUA_USERNAME> --docker-password=<AQUA_PASSWORD> --docker-email=<AQUA_EMAIL> -n aqua

kubectl create -f deploy/service_account.yaml -n aqua
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml -n aqua
