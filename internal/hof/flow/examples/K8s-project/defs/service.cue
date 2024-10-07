package defs

service: {
  apiVersion: "v1"
  kind:       "Service"
  metadata: {
    name:      "ingress-nginx-controller"
    namespace: "ingress-nginx"
  }
  spec: {
    type:     "NodePort"
    selector: {
      "app.kubernetes.io/name": "ingress-nginx"
    }
    ports: [{
      name:       "http"
      protocol:   "TCP"
      port:       80
      targetPort: 80
      nodePort:   30080
    }, {
      name:       "https"
      protocol:   "TCP"
      port:       443
      targetPort: 443
      nodePort:   30443
    }]
  }
}