package defs

#providers: {
    provider: {
        "aws": {
            region: "us-west-2"  // Changed to match the region in the deployment
        }
    }
    terraform: {
        required_providers: {
            aws: {
                source:  "hashicorp/aws"
                version: ">= 4.67.0"
            }
        }
    }
}

project_name: "flask-dynamodb-app"  // Updated project name

common: {   
    // Common configurations for the DynamoDB setup
    project_name: "flask-dynamodb-app"

    // DynamoDB table name
    table_name: "HelloWorldTable"

    // AWS region
    aws_region: "us-west-2"

    // Container repository
    container_repo: "registry.gitlab.com/flashresolve1/augur:v1"

    // Application port
    app_port: 5000

    // Environment
    environment: "Development"
}