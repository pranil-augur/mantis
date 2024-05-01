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

	_conn: data_source: "https://api.example.com/data?api_key=your_api_key"

	fetch: {
		@task(cueform.TerraformDataSourceTask)
		script: #common  
	}


	done: {
		@task(os.Stdout)
		text: "Data fetching and processing completed.\n"
	}
}
