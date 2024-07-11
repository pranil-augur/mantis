package middleware

import (
	"fmt"
	"regexp"
	"strings"

	"cuelang.org/go/cue"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
)

type Policy struct {
	val  cue.Value
	next hofcontext.Runner
}

func NewPolicy() *Policy {
	return &Policy{}
}

func applyPolicies(val cue.Value) error {
	// Directly look for the 'policy' attribute in the cue.Value
	attrs := val.Attributes(cue.ValueAttr)

	for _, attr := range attrs {
		if attr.Name() == "policy" {
			targetContent := attr.Contents()
			policyName := "_policy"

			// Lookup the policy using the extracted name
			policyVal := val.LookupPath(cue.ParsePath("policy"))
			fmt.Println("Policy Value:", policyVal)
			if policyVal.Err() != nil {
				return fmt.Errorf("failed to lookup policy value for %s: %v", policyName, policyVal.Err())
			}

			// Assuming the policy applies to the entire value
			if err := applyPolicy(policyVal, val, targetContent); err != nil {
				return fmt.Errorf("policy application failed: %v", err)
			}
			break // Assuming only one policy attribute is expected
		}
	}

	return nil
}

func applyPolicy(policy, target cue.Value, targetTag string) error {
	// Example policy application logic
	pattern, err := policy.LookupPath(cue.ParsePath("rule.pattern")).String()
	if err != nil {
		return fmt.Errorf("failed to extract policy rule: %v", err)
	}

	// Lookup the targetString in the target
	targetStringVal := target.LookupPath(cue.ParsePath(targetTag))
	if err != nil {
		return fmt.Errorf("failed to lookup target string '%s' in target: %v", targetTag, targetStringVal)
	}
	fmt.Println("Extracted Target String for Policy Evaluation:", targetStringVal)

	matched, err := regexp.MatchString(pattern, targetStringVal)
	if err != nil {
		return fmt.Errorf("regex match failed: %v", err)
	}

	if !matched {
		return fmt.Errorf("policy violation: value '%s' does not match pattern '%s'", targetStringVal, pattern)
	}

	return nil
}

// Helper function to extract the policy name from an annotation
func extractPolicyName(annotationText string) string {
	// Extract the policy name from the annotation text, e.g., "@policy(#cluster_name)" -> "cluster_name"
	// This is a placeholder function
	return strings.TrimPrefix(strings.TrimSuffix(annotationText, ")"), "@policy(#")
}

func (p *Policy) Run(ctx *hofcontext.Context) (results interface{}, err error) {
	// Apply policies to the configuration
	if err := applyPolicies(p.val); err != nil {
		return nil, err
	}

	// Execute the next runner in the chain
	result, err := p.next.Run(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func hasPolicy(val cue.Value, policyName string) bool {
	// Implement logic to check for the specified policy in the cue.Value
	// This is a placeholder function
	return true
}

func (p *Policy) Apply(ctx *hofcontext.Context, runner hofcontext.RunnerFunc) hofcontext.RunnerFunc {
	return func(val cue.Value) (hofcontext.Runner, error) {
		// Check if the value has the required policy attribute
		if hasPolicy(val, "cluster_name") {
			fmt.Println("Policy 'cluster_name' will be applied on:", val.Path())
		}

		next, err := runner(val)
		if err != nil {
			return nil, err
		}

		return &Policy{
			val:  val,
			next: next,
		}, nil
	}
}
