package defs

flaskDynamoDBDeployment: {
    apiVersion: "apps/v1"
    kind:       "Deployment"
    metadata: {
        name: "flask-dynamodb-app"
    }
    spec: {
        replicas: 2
        selector: {
            matchLabels: {
                app: "flask-dynamodb-app"
            }
        }
        template: {
            metadata: {
                labels: {
                    app: "flask-dynamodb-app"
                }
            }
            spec: {
                containers: [{
                    name:  "flask-dynamodb-app"
                    image: "\(common.container_repo):latest"
                    ports: [{
                        containerPort: 5000
                    }]
                    env: [
                        {
                            name: "AWS_ACCESS_KEY_ID"
                            valueFrom: {
                                secretKeyRef: {
                                    name: "aws-secret"
                                    key:  "aws-access-key-id"
                                }
                            }
                        },
                        {
                            name: "AWS_SECRET_ACCESS_KEY"
                            valueFrom: {
                                secretKeyRef: {
                                    name: "aws-secret"
                                    key:  "aws-secret-access-key"
                                }
                            }
                        },
                        {
                            name:  "AWS_DEFAULT_REGION"
                            value: "us-west-2"
                        },
                        {
                            name:  "DYNAMODB_TABLE_NAME"
                            value: "@var(table_name)"
                        }
                    ]
                }]
            }
        }
    }
}