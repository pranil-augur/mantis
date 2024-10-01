package schemas

#ModuleEC2: {
    source:  string
    version: string
    name:                   string
    ami:                    string
    instance_type:          string
    vpc_security_group_ids: [...string]
    subnet_id:              string
    rds_endpoint:           string
    user_data:              string
    tags:                   [string]:string
}