Kubernetes controller for webservers
-----------------------------------------------------------------------

## Installation:

1. Install helm
```bash
sudo curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
```
2. Add ingress-nginx repository into helm
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
```
3. Update helm
```bash
helm repo update
```
4. Install ingress-nginx
helm install quickstart ingress-nginx/ingress-nginx --set controller.service.type=NodePort --set controller.service.httpPort.nodePort=32526 --set controller.service.httpsPort.nodePort=30523

5. Install cert-manager
```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.16.1/cert-manager.yaml
```
6. Add cert manager resource object with let's encrypt
```bash
kubectl apply -f -
```
paste in this:
```bash
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
  namespace: cert-manager
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: your_email_address_here
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod
    # Enable the HTTP-01 challenge provider
    solvers:
    - http01:
        ingress:
          class: nginx
```
7. Install kubopwebdep into kubernetes
```bash
kubectl apply -f https://github.com/bartam1/kubopwebdep/releases/download/v1.0/kubopwebdep.yml
```

8. Add kubopwebdep resource object with your preferences
```bash
kubectl apply -f -
```
paste this:
```bash 
apiVersion: crd.bartam/v1
kind: WebInstall
metadata:
  name: webinstall-example
spec:
  replicas: 5
  host: "your-external-hostname.example"
  image: "nginx:latest"
```
