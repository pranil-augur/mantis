package defs

flaskDynamoDBService: {
    apiVersion: "v1"
    kind:       "Service"
    metadata: {
        name: "flask-dynamodb-service"
    }
    spec: {
        selector: {
            app: "flask-dynamodb-app"
        }
        ports: [{
            protocol:   "TCP"
            port:       80
            targetPort: 5000
        }]
        type: "LoadBalancer"
    }
}