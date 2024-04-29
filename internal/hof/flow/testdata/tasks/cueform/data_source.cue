package cueform

tasks: {
	@flow(data_source)

	_conn: data_source: "https://api.example.com/data?api_key=your_api_key"

	fetch: {
		@task(cueform.TerraformDataSourceTask)
        dataSourceName: "aws_availability_zone"
	}


	done: {
		@task(os.Stdout)
		text: "Data fetching and processing completed.\n"
	}
}
