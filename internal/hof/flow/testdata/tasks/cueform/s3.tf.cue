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

// Generalized policy definition
#Policy: {
	description: string
	rule: {
		pattern: string
		enforce: bool
	}
}

#s3_bucket_name: #Policy & {
	description: "Ensure EKS cluster names follow the required convention"
	rule: {
		pattern: "^eks-[a-zA-Z0-9]+-[a-zA-Z0-9]+$"
		enforce: true
	}
}

tasks: {
	@flow(s3_setup)

	setup: {
		@task(cueform.TF)
		script: #s3BucketConfig
	}

	done: {
		@task(os.Stdout)
		text: "S3 bucket setup completed.\`n"
	}
}
