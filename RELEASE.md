# sentinel-bot v1.0.0 — First Helm Release

- Helm chart with optional arch suffix
- Default args: ["sentinel-bot","sentinel-bot"]
- TELE_TOKEN via Secret (values.teleToken)
- Probes disabled by default

## Install
kubectl create namespace demo || true
kubectl -n demo create secret generic sentinel-bot --from-literal=token='<YOUR_TELE_TOKEN>'

helm install sentinel-bot https://github.com/mexxo-dvp/sentinel-bot/releases/download/v1.0.0/sentinel-bot-0.1.2.tgz \
  -n demo \
  --set image.repository=quay.io/paranoidlookup \
  --set image.tag=v1.0.0
