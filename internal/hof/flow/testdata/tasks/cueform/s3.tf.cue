package test

#providers: {
    provider: {
        "aws": {}
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

// S3 bucket configuration
#s3BucketConfig: {
	resource: {
		"aws_s3_bucket": {
			"otfork-sample-bucket": {
				bucket: "otfork-sample-bucket"
				tags: {
					Name:  string @runinject(tasks.get_azs_data.output.region)
					Environment: "dev"
				}
			}
		}
	}
}

#aws_availability_zones: {
    "data": {
        "aws_availability_zones": {
            "available": [
                {
                    "state": "available" 
                }
            ]
        }
    },
} 

tasks: {
	@flow(s3_setup)

	setup_providers: {
		@task(opentf.TF)
		config: #providers
	}

	get_azs_data: {
		@task(opentf.TF)
        dep: setup_providers
		config: #aws_availability_zones
		output: {
			region: "us-east1"
		}
	}

	setup_s3: {
		@task(opentf.TF)
		dep: [setup_providers, get_azs_data]
		config: #s3BucketConfig
	}

	done: {
		@task(os.Stdout)
		text: "S3 bucket setup completed.\n"
	}
}
