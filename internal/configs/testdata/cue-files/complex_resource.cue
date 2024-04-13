package resources

resource: {
    "aws_instance": {
        "example": {
            ami:           "ami-0c55b159cbfafe1f0"
            instance_type: "t2.micro"
            tags: {
                Name: "ExampleInstance"
            }
        }
    }
    "aws_vpc": {
        "main": {
            cidr_block: "10.0.0.0/16"
            tags: {
                Name: "main-network"
            }
        }
    }
}