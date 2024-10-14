# Mantis 

Mantis is an infrastructure as code (IaC) tool,
powered by [CUE](https://cuelang.org/)
and inspired by [OpenTofu](https://opentofu.org/) and [Helm](https://helm.sh/).

> [!IMPORTANT]
> Note that Mantis is under active development and is not yet ready for production use.
> The APIs and command-line interface may change in a backward incompatible manner.


## Mantis Vision
![](https://github.com/pranil-augur/mantis/blob/5db82db255a3e2af02288699af5a0af83d8a0cfd/mantis_vision.png)


### Mantis Demo
[![Mantis demo video](mantis_thumbnail.png)](https://www.loom.com/share/b8c48935df8f4752b305e64fc3bb3845)

### Key features of Mantis include:

**Infrastructure as Code (IaC)**: Mantis uses the CUE language for infrastructure descriptions, unifying Terraform modules, Kubernetes manifests, and other cloud-native tools into a single framework. It simplifies configuration management and enables reusable, version-controlled infrastructure blueprints. Mantis treats configuration as data.

**Package Management**: Mantis introduces a package management system for reusable infrastructure components by leveraging CUE modules. This enables consistent, versioned infrastructure configurations across environments and teams.

**Built-in Policies**: Mantis leverages CUE to allow teams to define and enforce security, compliance, and operational policies directly within configurations. Policies ensure that infrastructure adheres to standards before deployment, reducing misconfigurations and maintaining compliance. Mantis via CUE enables better collaboration between security and dev teams through efficient and granular policy reuse.  


### Quick install (Mac and Linux)
```
brew install pranil-augur/homebrew-mantis/mantis
```

**Documentation**
[Getting started](https://mantis.getaugur.ai/docs/getting_started/installation)

### Acknowledgements
As we look toward the future, we see tremendous growth in infrastructure demands, both in scale and complexity. Mantis builds on the strong shoulders of Terraform, OpenTofu and CUE and we deeply appreciate their work. We believe Mantis will play a critical role in helping organizations scale their infrastructure more efficiently as they prepare for the next wave of innovation, and that would benefit the overall automation ecosystem.

### Feedback and Community
We'd love to hear from you. Leave us a comment, join our slack or start a discussion 

**Slack channel**: [Join the community](https://mantiscommunity.slack.com/).

**License**: Mantis is currently under a proprietary license. We will soon move to a fair source license. (https://fair.io/licenses/).

