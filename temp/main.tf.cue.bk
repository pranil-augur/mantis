package main


// Populate the schema with the example data
terraform: {
        required_providers: {
            local: {
                source:  "hashicorp/local"
                version: "~> 2.0"
            }
        }
}

provider: {
     local: {}
}

resource: {
     local_file: {
         example: {
             content:  "1. Understand Cuelang Syntax and Semantics: Before making changes, ensure a deep understanding of how Cuelang structures configuration data."
             filename: "\\${path.module}/output.txt" 
         }
     }
}

