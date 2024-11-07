package infrastructure

// Define resources with references to other infrastructure components
resource: {
    webApp: {
        type: "kubernetes_deployment"
        name: "web-app"
        references: [{
            type: "SecurityGroup"
            id: "sg-frontend"
        }]
        ingress: [{
            from: "load-balancer"
            port: 443
        }]
        egress: [{
            to: "database"
            port: 5432
        }]
        infrastructure: {
            vpc: "vpc-123456"
            subnet: "subnet-frontend"
            securityGroups: ["sg-frontend"]
        }
    }

    backendAPI: {
        type: "kubernetes_deployment"
        name: "backend-api"
        references: [{
            type: "SecurityGroup"
            id: "sg-backend"
        }]
        ingress: [{
            from: "web-app"
            port: 8080
        }]
        egress: [{
            to: "database"
            port: 5432
        }]
        infrastructure: {
            vpc: "vpc-123456"
            subnet: "subnet-backend"
            securityGroups: ["sg-backend"]
        }
    }

    database: {
        type: "aws_rds_instance"
        name: "main-db"
        references: [{
            type: "SecurityGroup"
            id: "sg-database"
        }]
        ingress: [{
            from: "backend-api"
            port: 5432
        }]
        infrastructure: {
            vpc: "vpc-123456"
            subnet: "subnet-database"
            securityGroups: ["sg-database"]
        }
    }
}

