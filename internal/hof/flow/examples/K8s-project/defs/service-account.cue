package defs

serviceAccount: {
  apiVersion: "v1"
  kind:       "ServiceAccount"
  metadata: {
    name:      "ingress-nginx"
    namespace: "ingress-nginx"
  }
}
