package infrastructure

// Define a set of services with dependencies on one another
service: {
    frontend: {
        name: "frontend"
        dependencies: ["backend", "database"]
        ingress: [{
            from: "load-balancer"
            port: 443
        }]
        egress: [{
            to: "backend"
            port: 8080
        }]
    }

    backend: {
        name: "backend"
        dependencies: ["database", "cache"]
        ingress: [{
            from: "frontend"
            port: 8080
        }]
        egress: [{
            to: "database"
            port: 5432
        }]
    }

    database: {
        name: "database"
        dependencies: []
        ingress: [{
            from: "backend"
            port: 5432
        }]
        egress: []
    }

    cache: {
        name: "cache"
        dependencies: ["database"]
        ingress: []
        egress: []
    }
}

// Infrastructure dependencies for each service
infrastructure: {
    vpc: "vpc-123456"
    securityGroups: ["sg-frontend", "sg-backend", "sg-database", "sg-cache"]
    subnets: ["subnet-frontend", "subnet-backend", "subnet-database", "subnet-cache"]
}

