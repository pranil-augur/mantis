package defs

flaskRdsDeployment: {
    apiVersion: "apps/v1"
    kind:       "Deployment"
    metadata: {
        name:   "flask-rds-deployment"
        labels: {
            app: "flask-rds"
        }
    }
    spec: {
        replicas: 2
        selector: {
            matchLabels: {
                app: "flask-rds"
            }
        }
        template: {
            metadata: {
                labels: {
                    app: "flask-rds"
                }
            }
            spec: {
                containers: [{
                    name:  "flask-rds"
        	    image: "\(common.container_repo)"
                    ports: [{
                        containerPort: 80
                    }]
                    env: [
                        {
                            name:  "DB_HOST"
                            value: "@var(rds_endpoint)"
                        },
                        {
                            name:  "DB_NAME"
                            value: "\(common.db_name)"
                        },
                        {
                            name:  "DB_USER"
                            value: "\(common.db_username)"
                        },
                        {
                            name:  "DB_PASSWORD"
                            value: "\(common.db_password)"
                        }
                    ]
                    resources: {
                        limits: {
                            memory: "256Mi"
                            cpu:    "250m"
                        }
                        requests: {
                            memory: "128Mi"
                            cpu:    "80m"
                        }
                    }
                }]
            }
        }
    }
}

