# Mincraft with Crossplane

## Install Crossplane

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
```

```bash
helm upgrade -i crossplane crossplane-stable/crossplane \
--namespace crossplane-system \
--create-namespace
```

Working here with local credentials, this should be avoided in production

```bash
export AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id --profile private)
export AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key --profile private)
echo "[default]\naws_access_key_id = $AWS_ACCESS_KEY_ID\naws_secret_access_key = $AWS_SECRET_ACCESS_KEY\n" >aws-creds.conf
kubectl --namespace crossplane-system  create secret generic aws-creds --from-file creds=./aws-creds.conf
```

For DigitalOcean

```bash
export DIGITALOCEAN_ACCESS_TOKEN=xx
kubectl --namespace crossplane-system  create secret generic do-creds --from-literal=DIGITALOCEAN_TOKEN=$DIGITALOCEAN_TOKEN
```
