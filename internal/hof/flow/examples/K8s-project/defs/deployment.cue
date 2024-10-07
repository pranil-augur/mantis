package defs

deployment: {
  apiVersion: "apps/v1"
  kind:       "Deployment"
  metadata: {
    name:      "ingress-nginx-controller"
    namespace: "ingress-nginx"
    labels: {
      "app.kubernetes.io/name":     "ingress-nginx"
      "app.kubernetes.io/part-of":  "ingress-nginx"
    }
  }
  spec: {
    replicas: 1
    selector: matchLabels: {
      "app.kubernetes.io/name": "ingress-nginx"
    }
    template: {
      metadata: {
        labels: {
          "app.kubernetes.io/name": "ingress-nginx"
        }
      }
      spec: {
        serviceAccountName: "ingress-nginx"
        containers: [{
          name:  "controller"
          image: "registry.k8s.io/ingress-nginx/controller:v1.11.2"
          args: [
            "/nginx-ingress-controller",
            "--election-id=ingress-nginx-leader",
            "--controller-class=k8s.io/ingress-nginx",
            "--ingress-class=nginx",
            "--configmap=$(POD_NAMESPACE)/ingress-nginx-controller",
            "--watch-ingress-without-class=true",
          ]
          env: [
            {
              name: "POD_NAME"
              valueFrom: {
                fieldRef: {
                  fieldPath: "metadata.name"
                }
              }
            },
            {
              name: "POD_NAMESPACE"
              valueFrom: {
                fieldRef: {
                  fieldPath: "metadata.namespace"
                }
              }
            }
          ]
          ports: [{
            name:          "http"
            containerPort: 80
            hostPort:      80
            protocol:      "TCP"
          }, {
            name:          "https"
            containerPort: 443
            hostPort:      443
            protocol:      "TCP"
          }]
          livenessProbe: {
            httpGet: {
              path: "/healthz"
              port: 10254
            }
            initialDelaySeconds: 10
            periodSeconds:       10
          }
          readinessProbe: {
            httpGet: {
              path: "/healthz"
              port: 10254
            }
            initialDelaySeconds: 10
            periodSeconds:       10
          }
          resources: requests: {
            cpu:    "100m"
            memory: "90Mi"
          }
        }]
        nodeSelector: {
          "ingress-ready":      "true"
          "kubernetes.io/os":   "linux"
        }
        tolerations: [{
          key:      "node-role.kubernetes.io/master"
          operator: "Equal"
          effect:   "NoSchedule"
        }, {
          key:      "node-role.kubernetes.io/control-plane"
          operator: "Equal"
          effect:   "NoSchedule"
        }]
      }
    }
  }
}
