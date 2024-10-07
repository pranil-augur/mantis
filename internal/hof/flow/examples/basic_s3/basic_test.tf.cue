package test

#s3BucketConfig: {
	resource: {
		"aws_s3_bucket": {
			"otfork-sample-bucket": {
				bucket: "otfork-sample-bucket"
				tags: {
					Name:        string @runinject(region_name)
					Environment: "dev"
				}
			}
		}
	}
}
// #bucketRegion: {
// 	region: string
// }
tasks: {
	@flow(s3_setup)

	get_region: {
		@task(opentf.TF)
		// output: {
		region: "us-east-1"
		// }
		outputs: ["region"]
		config: {}
	}

	setup_s3: {
		@task(opentf.TF)
		dep:     get_region
		imports: {
			region_name: "tasks.get_region.outputs.region"
		}
		config: #s3BucketConfig
	}

	// done: {
	// 	@task(os.Stdout)
	// 	text: "S3 bucket setup completed.\n"
	// }
}