package main

variables : networking : {
  vpcID: "vpc-123abc"
  subnets: {
    "subnet-1": { CIDR: "10.0.1.0/24", public: true },
    "subnet-2": { CIDR: "10.0.2.0/24", public: false }
  }
}

