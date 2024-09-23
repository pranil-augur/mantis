package nested

install_mongodb: {
	@flow()
	create_user: {
		@task(os.Stdout)
		text: "In create user \n"
	} 

	create_secret: {
		dep: create_user
		@task(os.Stdout)
		text: "In create secret \n"
	} 

	install_secret: {
		dep: create_secret
		@task(os.Stdout)
		text: "In install secret \n"
	}
}