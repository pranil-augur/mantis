package services

// Define services within a service mesh with ingress and egress dependencies
service: {
    frontend: {
        name: "frontend"
        ingress: [{
            from: "load-balancer"
            port: 443
        }]
        egress: [{
            to: "backend-api"
            port: 8080
        }]
    }

    backendAPI: {
        name: "backend-api"
        ingress: [{
            from: "frontend"
            port: 8080
        }]
        egress: [{
            to: "payment-api"
            port: 3000
        }]
    }

    paymentAPI: {
        name: "payment-api"
        ingress: [{
            from: "backend-api"
            port: 3000
        }]
        egress: [{
            to: "billing-service"
            port: 9090
        }]
    }

    billingService: {
        name: "billing-service"
        ingress: [{
            from: "payment-api"
            port: 9090
        }]
        egress: []
    }
}

