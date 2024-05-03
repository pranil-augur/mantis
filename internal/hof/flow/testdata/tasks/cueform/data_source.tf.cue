package test

// Common configuration values
#common: {
	data: {
		"aws_availability_zones": {
			"available": {
				state: "available"
			}
		}
	}
}

tasks: {
	@flow(data_source)

	fetch: {
		@task(cueform.DS)
		script: #common  
		out: [string]: string | *{}
	}

	done: {
		@task(cueform.PrintObj)
		args: fetch.out 
	}
}
