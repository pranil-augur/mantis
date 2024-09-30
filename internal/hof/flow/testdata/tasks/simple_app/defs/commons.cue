package defs

#outputdef: {
    alias: string
    path:  [...string]
}

#TFTask: {
    @task(opentf.TF)
    config: _
    dep: _
    out: _
    outputs?: [...#outputdef]
    inputs?: _
}

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

inputs: _

common: {
    project_name: "static-website"
    vpc_cidr:     "10.0.0.0/16"
}

locals: {
    azs: ["us-east-2a"] //, "us-east-1b"]
    private_subnets: ["10.0.1.0/24", "10.0.2.0/24"]
    public_subnets:  ["10.0.101.0/24", "10.0.102.0/24"]
}