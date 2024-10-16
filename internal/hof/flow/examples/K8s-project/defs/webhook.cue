package defs

webhook: {
  apiVersion: "admissionregistration.k8s.io/v1"
  kind:       "ValidatingWebhookConfiguration"
  metadata: {
    name:      "ingress-nginx-admission"
    namespace: "ingress-nginx"
  }
  webhooks: [{
    name: "validate.nginx.ingress.kubernetes.io"
    clientConfig: {
      service: {
        name:      "ingress-nginx-controller-admission"
        namespace: "ingress-nginx"
        path:      "/networking/v1/ingresses"
      }
      caBundle: ""
    }
    rules: [{
      apiGroups:   ["networking.k8s.io"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["ingresses"]
    }]
    admissionReviewVersions: ["v1"]
    sideEffects: "None"
    timeoutSeconds: 10
  }]
}