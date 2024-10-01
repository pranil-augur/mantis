package main

import (
    "augur.ai/simple-app/defs"
)

setup_s3_flow: {
    // The flow annotation is used to indicate a flow to mantis
    @flow(setup_s3)

    setup_providers: {
        @task(opentf.TFProviders)
        config: defs.#providers
    }

    setup_s3: {
        @task(opentf.TF)
        dep: setup_providers
        config: {
            resource: {
                aws_s3_bucket: {
                    simple_app_bucket: {
                        bucket: "\(defs.common.project_name)-bucket"
                        tags: {
                            Name:        "\(defs.common.project_name)-bucket"
                            Environment: "dev"
                            Project:     defs.common.project_name
                        }
                    }
                }
            }
        }
        
        outputs: [{
            alias: "bucket_id"
            path: ["aws_s3_bucket", "simple_app_bucket", "id"]
        }, {
            alias: "bucket_arn"
            path: ["aws_s3_bucket", "simple_app_bucket", "arn"]
        }]
    }

    print_values: {
        @task(opentf.TF)
        dep: setup_s3
        inputs: {
            bucket_id: "tasks.setup_s3.outputs.bucket_id"
        }
        config: {
            output: {
                bucket_info: {
                    value: [string] @inject(bucket_id) 
                }
            }
        }
    }
}

