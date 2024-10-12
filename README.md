# Mantis 

Mantis is a infrastructure as code (IaC) tool,
powered by [CUE](https://cuelang.org/)
and inspired by [OpenTofu](https://opentofu.org/) and [Helm](https://helm.sh/).

> [!IMPORTANT]
> Note that Mantis is under active development and is not yet ready for production use.
> The APIs and command-line interface may change in a backward incompatible manner.


## Mantis Vision
![](https://github.com/pranil-augur/mantis/blob/5db82db255a3e2af02288699af5a0af83d8a0cfd/mantis_vision.png)

The key features of Mantis are:

**Infrastructure as Code (IaC)**: Mantis uses the CUE language for infrastructure descriptions, unifying Terraform modules, Kubernetes manifests, and other cloud-native tools into a single framework. It simplifies configuration management and enables reusable, version-controlled infrastructure blueprints. Mantis treats configuration as data.

**Package Management**: Mantis introduces a package management system for reusable infrastructure components, including CUE modules, Terraform modules, and Kubernetes manifests. This ensures consistent, versioned infrastructure across environments and teams.

**Built-in Policies**: Mantis leverages CUE to allow teams to define and enforce security, compliance, and operational policies directly within configurations. Policies ensure that infrastructure adheres to standards before deployment, reducing misconfigurations and maintaining compliance.


**License**: Mantis is currently under a proprietary license. We will soon move to a fair source license. (https://fair.io/licenses/)


### Community
**Slack channel**: [Join the community](https://mantiscommunity.slack.com/)

**Documentation**: [Getting started](https://mantis.getaugur.ai/docs/getting_started/installation)
