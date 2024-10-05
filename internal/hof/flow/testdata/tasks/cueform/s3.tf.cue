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
					Name:        _  | *null @var(available_zones)
					Environment: "dev"
				}
			}
		}
	}
}

// S3 bucket configuration
#s3BucketConfig1: {
	resource: {
		"aws_s3_bucket": {
			"otfork-sample-bucket-1": {
				bucket: "otfork-sample-bucket-1"
				tags: {
					Name:        _  | *null @var(available_zones)
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
				},
			]
		}
	}
}

tasks: {
	@flow(s3_setup)
	setup_providers: {
		@task(opentf.TF)
		config: #providers
	}

	get_azs_data: {
		@task(opentf.TF)
		dep:    setup_providers
		config: #aws_availability_zones
		exports: [{
			alias: "available_zones"
			path:  ".data.aws_availability_zones.available.id"
		}]
	}

	local_eval: {
		@task(mantis.Evaluate)
		exports: [{	
			cueexpr: "strings.Split(get_azs_data.exports[0].alias, \"-\")[0]"
			var: "first_az"
		}]
	}

	setup_s3: {
		@task(opentf.TF)
		dep: [setup_providers, get_azs_data]
		config: #s3BucketConfig
	}
	
	setup_s3_1: {
		@task(opentf.TF)
		dep: [setup_s3]
		config: #s3BucketConfig1
	}
}
