# kbot v1.0.3 — First Helm Release

- Helm chart with optional arch suffix
- Default args: ["kbot","kbot"]
- TELE_TOKEN via Secret (values.teleToken)
- Probes disabled by default

## Install
kubectl create namespace demo || true
kubectl -n demo create secret generic kbot --from-literal=token='<YOUR_TELE_TOKEN>'

helm install kbot https://github.com/mexxo-dvp/kbot/releases/download/v1.0.3/kbot-0.1.1.tgz \
  -n demo \
  --set image.repository=quay.io/paranoidlookup \
  --set image.tag=v1.0.3
