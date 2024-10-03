package schemas

#ModuleEC2: {
	source:        string
	version:       string
	name:          string
	ami:           string
	instance_type: string
	vpc_security_group_ids: [...string]
	subnet_id:              string | *null  
	rds_endpoint:           string | *null
	user_data:              string
	tags: [string]: string
}
