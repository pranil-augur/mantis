# Mantis 

Mantis is a converged infrastructure as code (IaC) tool that unifies Terraform modules, Helm charts, and other cloud-native tools into a single framework. 

The key features of Mantis are:

**Infrastructure as Code (IaC)**: Mantis uses the CUE language for infrastructure descriptions, unifying Terraform modules, Helm charts, and other cloud-native tools into a single framework. It simplifies configuration management and enables reusable, version-controlled infrastructure blueprints. Mantis treats configuration as data.

**Execution Plans**: Mantis generates execution plans before making changes, providing clear visibility into upcoming infrastructure operations, helping avoid unexpected changes, and ensuring safe updates.

**Package Management**: Mantis introduces a package management system for reusable infrastructure components, including CUE modules, Terraform modules, and Helm charts. This ensures consistent, versioned infrastructure across environments and teams.

**Built-in Policies**: Mantis leverages CUE to allow teams to define and enforce security, compliance, and operational policies directly within configurations. Policies ensure that infrastructure adheres to standards before deployment, reducing misconfigurations and maintaining compliance.

**Resource Graph**: Mantis builds a dependency graph of resources, parallelizing operations where possible to improve efficiency. This provides a clear view of resource dependencies and speeds up deployments.

### Project Status
Mantis is currently in the alpha stage. We are actively developing and testing the project, and we welcome feedback from the community. We **strongly advise against** using Mantis to manage cloud resources in production at this stage.

**License**: Mantis is currently under a proprietary license. We are evaluating other license options to make Mantis more accessible. We'd love to hear your thoughts on this.


### Community
**Slack channel**: [Join the community](https://mantiscommunity.slack.com/)

**Documentation**: [Getting started](https://getaugur.ai/docs/introduction/what_is_mantis)


### Known Limitations 
**Dependency management**

1. Currently, both Mantis destroy and apply are executed in the same direction. Ideally, destroy should run in the reverse direction of creation ie dependent resources should be destroyed first. 
2. Mantis does runtime variable injection, that's not visible to the path dependency calculator of the cuelang/tool/flow library. This can lead to incorrect dependency graph calculation. The current workaround is explicitly set the dep attribute to specify the dependencies.

**Backend support**   

1. Mantis currently supports only the local backend. Other backends may work but are not tested.

**Built-in functions**

1. CUE lang supports standard [out-of-the-box](https://cuetorials.com/overview/standard-library/) functions, however open tofu and terraform support other IaC domain-specific built-ins as listed on the [open tofu site](https://opentofu.org/docs/language/functions/)

**Other Limitations**

Please report any other limitations you find to the Mantis team. You can use the Github issue tracker for this.

**Plan to address issue**

We will work with the CUE team to support these before the GA release.
