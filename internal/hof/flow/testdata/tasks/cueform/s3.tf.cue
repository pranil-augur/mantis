package test

// S3 bucket configuration
#s3BucketConfig: {
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
	resource: {
		"aws_s3_bucket": {
			"otfork-sample-bucket": {
				bucket: "otfork-sample-bucket"
				tags: {
					Name:        "ot-fork"
					Environment: "dev"
				}
			}
		}
	}
}

tasks: {
	@flow(s3_setup)

	setup: {
		@task(cueform.TerraformDataSourceTask)
		script: #s3BucketConfig
	}

	done: {
		@task(os.Stdout)
		text: "S3 bucket setup completed.\n"
	}
}