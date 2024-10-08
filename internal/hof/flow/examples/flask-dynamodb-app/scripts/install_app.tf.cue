package main

import (
    "augur.ai/rds-flask-app/defs"
)

deploy_flask_dynamodb: {
    @flow(deploy_flask_dynamodb)

    setup_tf_providers: {
        @task(mantis.core.TF)
        config: defs.#providers 
    }

    create_dynamodb_table: {
        @task(mantis.core.TF)
        dep: [setup_tf_providers]
        config: resource: aws_dynamodb_table: hello_world_table: {
            name:         "HelloWorldTable"
            billing_mode: "PAY_PER_REQUEST"
            hash_key:     "ID"
            attribute: [{
                name: "ID"
                type: "S"
            }]
            tags: {
                Name:        "HelloWorldTable"
                Environment: "Development"
            }
        }
        exports: [{
            path: ".aws_dynamodb_table.hello_world_table.name"
            var:  "table_name"
        }, {
            path: ".aws_dynamodb_table.hello_world_table.arn"
            var:  "table_arn"
        }]
    }

    deploy_flask_app: {
        @task(mantis.core.K8s)
        dep: [create_dynamodb_table]
        config: defs.flaskDynamoDBDeployment 
    }
}