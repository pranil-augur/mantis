# **Mantis**

🚀 AI-first queryable Infrastructure as Code tool that is an alternative to Terraform and Helm

**Introduction**

Mantis is a next-generation Infrastructure as Code (IaC) tool that reimagines how we manage cloud and Kubernetes resources. Built as a fork of OpenTofu and powered by CUE, Mantis combines the best of Terraform and Helm while solving their limitations.

### **Key Features**

* **Unified Configuration**: Single tool to replace both Terraform and Helm workflows  
* **Task-Centric State Management**: Unlike Terraform's global state, Mantis manages state at the task level, eliminating lock contentions and speeding up deployments  
* **AI-first:** Mantis treats configuration as code and applies Gen-AI to generate, validate, query and visualize configuration and config changes  
* **Built-in Policy Engine**: Define and enforce security, compliance, and operational policies

⚠️ **Note**: Mantis is under active development. APIs and CLI interfaces may change.

## **Installation**

### **Prerequisites**

* Basic understanding of IaC concepts  
* Familiarity with Terraform or Helm (helpful but not required)

### **Quick Install**

**MacOS/Linux**

```bash  
brew install pranil-augur/homebrew-mantis/mantis
```

For other platforms and methods, see our detailed installation guide.

## **Usage**

### **Basic Example \- Install a K8s based Flask app**

Deploy a cloud-native Flask application integrated with AWS RDS and managed through Kubernetes. We'll walk through the structure of the example code, how the tasks are broken down, and how the CUE-based modules simplify reusable infrastructure components.

Let’s dive into the file structure and flow that powers this deployment.

Project File Structure Below is the structure of the example deployment:

```bash
tree -L 2
.
├── cue.mod
│   └── module.cue           # Defines the module name and CUE version
├── defs
│   ├── deployment.cue       # Deployment configurations for the Flask app
│   ├── rds.cue              # RDS database configurations
│   ├── variables.cue        # Variable definitions and inputs
│   └── providers.cue        # Providers configuration
└── install_flask_app.tf.cue  # Main Mantis flow for deploying the app
```

This file structure reflects how Mantis organizes infrastructure code using modular and reusable CUE configurations. Let's look at what each file and directory does.

#### **1\. Main Flow (install\_flask\_app.tf.cue)[​](https://getmantis.ai/blog/mantis_application_install#1-main-flow-install_flask_apptfcue)**

Main Flow Code

The main flow file orchestrates the entire deployment process by:

* Importing and using the definitions from defs/  
* Defining task dependencies and execution order  
* Managing state and variable passing between tasks  
* Coordinating both AWS and Kubernetes resources

The core syntax of the main flow is:

```cue
deploy_flask_rds: {
   @flow(deploy_flask_rds)
   # Define the tasks that make up the flow
   task_1: {
       @task(mantis.core.TF) // Terraform task
       ...
   }
   task_2: {
       @task(mantis.core.TF)
       dep: [task_1] // Define task dependencies
       ...
   }
   task_3: {
       @task(mantis.core.K8s) // Kubernetes task
       dep: [task_1, task_2] // Define task dependencies
       ...
   }
}
```

#### **2\. cue.mod/module.cue[​](https://getmantis.ai/blog/mantis_application_install#2-cuemodmodulecue)**

This file defines the module name and CUE language version being used for the project. It also allows dependencies to be managed across the project.

```cue
module: "augur.ai/rds-flask-app"
language: {
   version: "v0.10.0"
}

// Define the dependencies for the project
dependencies: [
   "abc.xyz.com/module1",
   "abc.xyz.com/module2",
]
```

Purpose: This ensures the project remains compatible across various CUE versions and clearly identifies the module for import across multiple flows.

#### **3\. defs Directory[​](https://getmantis.ai/blog/mantis_application_install#3-defs-directory)**

* defs/deployment.cue  
* defs/variables.cue  
* defs/providers.cue  
* defs/rds.cue

```cue
package defs

flaskRdsDeployment: {
apiVersion: "apps/v1"
kind:       "Deployment"
metadata: {
   name:   "flask-rds-deployment"
   labels: {
       app: "flask-rds"
   }
}
spec: {
   replicas: 2
   selector: {
       matchLabels: {
           app: "flask-rds"
       }
   }
   template: {
       metadata: {
           labels: {
               app: "flask-rds"
           }
       }
       spec: {
           containers: [{
               name:  "flask-rds"
   	    image: "\(common.container_repo)"
               ports: [{
                   containerPort: 80
               }]
               env: [
                   {
                       name:  "DB_HOST"
                       value: "@var(rds_endpoint)"
                   },
                   {
                       name:  "DB_NAME"
                       value: "\(common.db_name)"
                   },
                   {
                       name:  "DB_USER"
                       value: "\(common.db_username)"
                   },
                   {
                       name:  "DB_PASSWORD"
                       value: "\(common.db_password)"
                   }
               ]
               resources: {
                   limits: {
                       memory: "256Mi"
                       cpu:    "250m"
                   }
                   requests: {
                       memory: "128Mi"
                       cpu:    "80m"
                   }
               }
           }]
           }
       }
   }
```
## **Demo video**
[Introduction to Mantis](https://www.loom.com/share/b8c48935df8f4752b305e64fc3bb3845)

## **Documentation**

### **Core Concepts**

* [Tasks](https://getmantis.ai/docs/key_concepts/flows/tasks) & [flows](https://getmantis.ai/docs/key_concepts/flows/flow_overview)

### **Guides**

* [Getting Started](https://getmantis.ai/docs/getting_started/installation)  
* [Migrating from Terraform](Coming soon)  
* [Migrating from Helm](Coming soon)
* [Codifying Cloud Infrastructure](Coming soon)

## **Contributing**

* The easiest way to contribute is to pick an issue with the `good first issue` tag 💪. Read the contribution guidelines here.  
* Submit your bug reports and feature requests [here](https://github.com/pranil-augur/mantis/issues)

---

## **Community**

* Join our growing community around the world, for help, ideas, and discussions   
  * Discord  
  * Twitter

