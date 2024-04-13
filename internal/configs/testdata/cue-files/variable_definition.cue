// testdata/cue-files/variable_definition.cue
variable: {
    instance_type : {
        description : "The instance type of the EC2 instance"
        type        : "string"
        default     : "t2.micro"
    }
}