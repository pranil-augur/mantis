package main

// Define a generic structure for a Terraform provider
#Provider: {
    "provider": {
        aws: {
            version: string
        }
    }
}

// Define a generic structure for Terraform S3 bucket resource
#S3Bucket: {
    "resource": {
        "aws_s3_bucket": {
            [string]: {
                bucket: string
                tags: [string]: string
            }
        }
    }
}


// AWS provider configuration with version details
awsProvider: #Provider & {
    "provider": {
        aws: {
            version: ">= 4.67.0"
        }
    }
}

// S3 bucket configuration with specific bucket name and tags
s3Bucket: #S3Bucket & {
    "resource": {
        "aws_s3_bucket": {
            "otfork-sample-bucket": {
                bucket: "otfork-sample-bucket"
                tags: {
                    "Name":        "ot-fork"
                    "Environment": "dev"
                }
            }
        }
    }
}

// Combine provider and resource configurations into a blueprint
cueform: {
    provider: awsProvider.provider
    resource: s3Bucket.resource
}

