# Mantis 

[Mantis] is a infrastructure as code (IaC) tool,
powered by [CUE](https://cuelang.org/)
and inspired by [OpenTofu](https://opentofu.org/) and [Helm](https://helm.sh/).


> [!IMPORTANT]
> Note that Mantis is under active development and is not yet ready for production use.
> The APIs and command-line interface may change in a backwards incompatible manner.


The key features of Mantis are:

**Infrastructure as Code (IaC)**: Mantis uses the CUE language for infrastructure descriptions, unifying Terraform modules, Kubernetes manifests, and other cloud-native tools into a single framework. It simplifies configuration management and enables reusable, version-controlled infrastructure blueprints. Mantis treats configuration as data.

**Execution Plans**: Mantis generates execution plans before making changes, providing clear visibility into upcoming infrastructure operations, helping avoid unexpected changes, and ensuring safe updates.

**Package Management**: Mantis introduces a package management system for reusable infrastructure components, including CUE modules, Terraform modules, and Helm charts. This ensures consistent, versioned infrastructure across environments and teams.

**Built-in Policies**: Mantis leverages CUE to allow teams to define and enforce security, compliance, and operational policies directly within configurations. Policies ensure that infrastructure adheres to standards before deployment, reducing misconfigurations and maintaining compliance.


**License**: Mantis is currently under a proprietary license. We are evaluating other license options to make Mantis more accessible. We'd love to hear your thoughts on this.


### Community
**Slack channel**: [Join the community](https://mantiscommunity.slack.com/)

**Documentation**: [Getting started](https://getaugur.ai/docs/introduction/what_is_mantis)
