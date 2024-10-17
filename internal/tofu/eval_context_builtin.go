// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tofu

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/checks"
	"github.com/opentofu/opentofu/internal/configs/configschema"
	"github.com/opentofu/opentofu/internal/encryption"
	hofcontext "github.com/opentofu/opentofu/internal/hof/flow/context"
	"github.com/opentofu/opentofu/internal/instances"
	"github.com/opentofu/opentofu/internal/lang"
	"github.com/opentofu/opentofu/internal/plans"
	"github.com/opentofu/opentofu/internal/providers"
	"github.com/opentofu/opentofu/internal/provisioners"
	"github.com/opentofu/opentofu/internal/refactoring"
	"github.com/opentofu/opentofu/internal/states"
	"github.com/opentofu/opentofu/internal/tfdiags"
	"github.com/opentofu/opentofu/version"
)

// BuiltinEvalContext is an EvalContext implementation that is used by
// OpenTofu by default.
type BuiltinEvalContext struct {
	// StopContext is the context used to track whether we're complete
	StopContext context.Context

	// PathValue is the Path that this context is operating within.
	PathValue addrs.ModuleInstance

	// pathSet indicates that this context was explicitly created for a
	// specific path, and can be safely used for evaluation. This lets us
	// differentiate between PathValue being unset, and the zero value which is
	// equivalent to RootModuleInstance.  Path and Evaluation methods will
	// panic if this is not set.
	pathSet bool

	// Evaluator is used for evaluating expressions within the scope of this
	// eval context.
	Evaluator *Evaluator

	// VariableValues contains the variable values across all modules. This
	// structure is shared across the entire containing context, and so it
	// may be accessed only when holding VariableValuesLock.
	// The keys of the first level of VariableValues are the string
	// representations of addrs.ModuleInstance values. The second-level keys
	// are variable names within each module instance.
	VariableValues     map[string]map[string]cty.Value
	VariableValuesLock *sync.Mutex

	// Plugins is a library of plugin components (providers and provisioners)
	// available for use during a graph walk.
	Plugins *contextPlugins

	Hooks                 []Hook
	InputValue            UIInput
	ProviderCache         map[string]providers.Interface
	ProviderInputConfig   map[string]map[string]cty.Value
	ProviderLock          *sync.Mutex
	ProvisionerCache      map[string]provisioners.Interface
	ProvisionerLock       *sync.Mutex
	FunctionCache         *ProviderFunctions
	FunctionLock          sync.Mutex
	ChangesValue          *plans.ChangesSync
	StateValue            *states.SyncState
	ChecksValue           *checks.State
	RefreshStateValue     *states.SyncState
	PrevRunStateValue     *states.SyncState
	InstanceExpanderValue *instances.Expander
	MoveResultsValue      refactoring.MoveResults
	ImportResolverValue   *ImportResolver
	Encryption            encryption.Encryption
	TfContext             *hofcontext.TFContext
}

// BuiltinEvalContext implements EvalContext
var _ EvalContext = (*BuiltinEvalContext)(nil)

func (ctx *BuiltinEvalContext) WithPath(path addrs.ModuleInstance) EvalContext {
	newCtx := *ctx
	newCtx.pathSet = true
	newCtx.PathValue = path
	newCtx.FunctionCache = nil
	newCtx.FunctionLock = sync.Mutex{}
	return &newCtx
}

func (ctx *BuiltinEvalContext) UpdateHofCtxVariables(key string, vars cty.Value) *sync.Map {
	// Ensure TfContext is initialized
	if ctx.TfContext == nil {
		ctx.TfContext = &hofcontext.TFContext{}
	}

	// Initialize ParsedVariables if it's nil
	if ctx.TfContext.ParsedVariables == nil {
		ctx.TfContext.ParsedVariables = &sync.Map{}
	}

	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '.' || r == '[' || r == ']'
	})

	current := ctx.TfContext.ParsedVariables
	for i, part := range parts {
		if i == len(parts)-1 {
			// At the last part, assign the new value
			goMap, err := convertCtyValueMapToGoMap(vars.AsValueMap())
			if err != nil {
				log.Printf("[ERROR] Failed to convert value for key %s: %v", key, err)
				return ctx.TfContext.ParsedVariables
			}
			nestedSyncMap, err := convertToSyncMap(goMap)
			if err != nil {
				log.Printf("[ERROR] Failed to convert to sync.Map for key %s: %v", key, err)
				return ctx.TfContext.ParsedVariables
			}
			current.Store(part, nestedSyncMap)
		} else {
			var nextLevel *sync.Map
			if value, ok := current.Load(part); ok {
				switch v := value.(type) {
				case *sync.Map:
					nextLevel = v
				case []interface{}:
					index, err := strconv.Atoi(parts[i+1])
					if err != nil {
						// Handle error: parts[i+1] is not a valid integer index
						return ctx.TfContext.ParsedVariables
					}
					// Ensure the slice is large enough
					for len(v) <= index {
						v = append(v, &sync.Map{})
					}
					current.Store(part, v)
					nextLevel = v[index].(*sync.Map)
					i++ // Skip the next part as we've used it as an index
				default:
					// If it's neither a map nor a slice, overwrite with a new sync.Map
					nextLevel = &sync.Map{}
					current.Store(part, nextLevel)
				}
			} else {
				// If it doesn't exist, create a new sync.Map
				nextLevel = &sync.Map{}
				current.Store(part, nextLevel)
			}
			current = nextLevel
		}
	}

	return ctx.TfContext.ParsedVariables
}

func convertCtyValueMapToGoMap(ctyMap map[string]cty.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range ctyMap {
		switch {
		case value.IsNull():
			result[key] = nil
		case value.Type() == cty.String:
			result[key] = value.AsString()
		case value.Type() == cty.Number:
			if v, accuracy := value.AsBigFloat().Float64(); accuracy == big.Exact {
				result[key] = v
			} else {
				return nil, fmt.Errorf("unable to convert number to float64 for key %s", key)
			}
		case value.Type() == cty.Bool:
			result[key] = value.True()
		case value.Type().IsMapType() || value.Type().IsObjectType():
			nestedMap, err := convertCtyValueMapToGoMap(value.AsValueMap())
			if err != nil {
				return nil, fmt.Errorf("error in nested map for key %s: %w", key, err)
			}
			result[key] = nestedMap
		case value.Type().IsListType() || value.Type().IsTupleType():
			slice := value.AsValueSlice()
			list := make([]interface{}, len(slice))
			for i, v := range slice {
				converted, err := convertCtyValueToInterface(v)
				if err != nil {
					return nil, fmt.Errorf("error in list element %d for key %s: %w", i, key, err)
				}
				list[i] = converted
			}
			result[key] = list
		default:
			return nil, fmt.Errorf("unsupported type for key %s: %s", key, value.Type().FriendlyName())
		}
	}

	return result, nil
}

func convertCtyValueToInterface(value cty.Value) (interface{}, error) {
	switch {
	case value.IsNull():
		return nil, nil
	case value.Type() == cty.String:
		return value.AsString(), nil
	case value.Type() == cty.Number:
		if v, accuracy := value.AsBigFloat().Float64(); accuracy == big.Exact {
			return v, nil
		}
		return nil, fmt.Errorf("unable to convert number to float64")
	case value.Type() == cty.Bool:
		return value.True(), nil
	case value.Type().IsMapType() || value.Type().IsObjectType():
		return convertCtyValueMapToGoMap(value.AsValueMap())
	case value.Type().IsListType() || value.Type().IsTupleType():
		slice := value.AsValueSlice()
		list := make([]interface{}, len(slice))
		for i, v := range slice {
			converted, err := convertCtyValueToInterface(v)
			if err != nil {
				return nil, fmt.Errorf("error in list element %d: %w", i, err)
			}
			list[i] = converted
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", value.Type().FriendlyName())
	}
}

func convertToSyncMap(input map[string]interface{}) (*sync.Map, error) {
	newSyncMap := &sync.Map{}
	for k, v := range input {
		switch nested := v.(type) {
		case map[string]interface{}:
			nestedSyncMap, err := convertToSyncMap(nested)
			if err != nil {
				return nil, fmt.Errorf("error in nested map for key %s: %w", k, err)
			}
			newSyncMap.Store(k, nestedSyncMap)
		default:
			newSyncMap.Store(k, v)
		}
	}
	return newSyncMap, nil
}

func (ctx *BuiltinEvalContext) Stopped() <-chan struct{} {
	// This can happen during tests. During tests, we just block forever.
	if ctx.StopContext == nil {
		return nil
	}

	return ctx.StopContext.Done()
}

func (ctx *BuiltinEvalContext) Hook(fn func(Hook) (HookAction, error)) error {
	for _, h := range ctx.Hooks {
		action, err := fn(h)
		if err != nil {
			return err
		}

		switch action {
		case HookActionContinue:
			continue
		case HookActionHalt:
			// Return an early exit error to trigger an early exit
			log.Printf("[WARN] Early exit triggered by hook: %T", h)
			return nil
		}
	}

	return nil
}

func (ctx *BuiltinEvalContext) Input() UIInput {
	return ctx.InputValue
}

func (ctx *BuiltinEvalContext) InitProvider(addr addrs.AbsProviderConfig) (providers.Interface, error) {
	// If we already initialized, it is an error
	if p := ctx.Provider(addr); p != nil {
		return nil, fmt.Errorf("%s is already initialized", addr)
	}

	// Warning: make sure to acquire these locks AFTER the call to Provider
	// above, since it also acquires locks.
	ctx.ProviderLock.Lock()
	defer ctx.ProviderLock.Unlock()

	key := addr.String()

	p, err := ctx.Plugins.NewProviderInstance(addr.Provider)
	if err != nil {
		return nil, err
	}

	log.Printf("[TRACE] BuiltinEvalContext: Initialized %q provider for %s", addr.String(), addr)
	ctx.ProviderCache[key] = p

	return p, nil
}

func (ctx *BuiltinEvalContext) Provider(addr addrs.AbsProviderConfig) providers.Interface {
	ctx.ProviderLock.Lock()
	defer ctx.ProviderLock.Unlock()

	return ctx.ProviderCache[addr.String()]
}

func (ctx *BuiltinEvalContext) ProviderSchema(addr addrs.AbsProviderConfig) (providers.ProviderSchema, error) {
	return ctx.Plugins.ProviderSchema(addr.Provider)
}

func (ctx *BuiltinEvalContext) CloseProvider(addr addrs.AbsProviderConfig) error {
	ctx.ProviderLock.Lock()
	defer ctx.ProviderLock.Unlock()

	key := addr.String()
	provider := ctx.ProviderCache[key]
	if provider != nil {
		delete(ctx.ProviderCache, key)
		return provider.Close()
	}

	return nil
}

func (ctx *BuiltinEvalContext) ConfigureProvider(addr addrs.AbsProviderConfig, cfg cty.Value) tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics
	if !addr.Module.Equal(ctx.Path().Module()) {
		// This indicates incorrect use of ConfigureProvider: it should be used
		// only from the module that the provider configuration belongs to.
		panic(fmt.Sprintf("%s configured by wrong module %s", addr, ctx.Path()))
	}

	p := ctx.Provider(addr)
	if p == nil {
		diags = diags.Append(fmt.Errorf("%s not initialized", addr))
		return diags
	}

	req := providers.ConfigureProviderRequest{
		TerraformVersion: version.String(),
		Config:           cfg,
	}

	resp := p.ConfigureProvider(req)
	return resp.Diagnostics
}

func (ctx *BuiltinEvalContext) ProviderInput(pc addrs.AbsProviderConfig) map[string]cty.Value {
	ctx.ProviderLock.Lock()
	defer ctx.ProviderLock.Unlock()

	if !pc.Module.Equal(ctx.Path().Module()) {
		// This indicates incorrect use of InitProvider: it should be used
		// only from the module that the provider configuration belongs to.
		panic(fmt.Sprintf("%s initialized by wrong module %s", pc, ctx.Path()))
	}

	if !ctx.Path().IsRoot() {
		// Only root module provider configurations can have input.
		return nil
	}

	return ctx.ProviderInputConfig[pc.String()]
}

func (ctx *BuiltinEvalContext) SetProviderInput(pc addrs.AbsProviderConfig, c map[string]cty.Value) {
	absProvider := pc
	if !pc.Module.IsRoot() {
		// Only root module provider configurations can have input.
		log.Printf("[WARN] BuiltinEvalContext: attempt to SetProviderInput for non-root module")
		return
	}

	// Save the configuration
	ctx.ProviderLock.Lock()
	ctx.ProviderInputConfig[absProvider.String()] = c
	ctx.ProviderLock.Unlock()
}

func (ctx *BuiltinEvalContext) Provisioner(n string) (provisioners.Interface, error) {
	ctx.ProvisionerLock.Lock()
	defer ctx.ProvisionerLock.Unlock()

	p, ok := ctx.ProvisionerCache[n]
	if !ok {
		var err error
		p, err = ctx.Plugins.NewProvisionerInstance(n)
		if err != nil {
			return nil, err
		}

		ctx.ProvisionerCache[n] = p
	}

	return p, nil
}

func (ctx *BuiltinEvalContext) ProvisionerSchema(n string) (*configschema.Block, error) {
	return ctx.Plugins.ProvisionerSchema(n)
}

func (ctx *BuiltinEvalContext) CloseProvisioners() error {
	var diags tfdiags.Diagnostics
	ctx.ProvisionerLock.Lock()
	defer ctx.ProvisionerLock.Unlock()

	for name, prov := range ctx.ProvisionerCache {
		err := prov.Close()
		if err != nil {
			diags = diags.Append(fmt.Errorf("provisioner.Close %s: %w", name, err))
		}
	}

	return diags.Err()
}

func (ctx *BuiltinEvalContext) EvaluateBlock(body hcl.Body, schema *configschema.Block, self addrs.Referenceable, keyData InstanceKeyEvalData) (cty.Value, hcl.Body, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	scope := ctx.EvaluationScope(self, nil, keyData)
	body, evalDiags := scope.ExpandBlock(body, schema)
	diags = diags.Append(evalDiags)
	val, evalDiags := scope.EvalBlock(body, schema)
	diags = diags.Append(evalDiags)
	return val, body, diags
}

func (ctx *BuiltinEvalContext) EvaluateExpr(expr hcl.Expression, wantType cty.Type, self addrs.Referenceable) (cty.Value, tfdiags.Diagnostics) {
	scope := ctx.EvaluationScope(self, nil, EvalDataForNoInstanceKey)
	return scope.EvalExpr(expr, wantType)
}

func (ctx *BuiltinEvalContext) EvaluateReplaceTriggeredBy(expr hcl.Expression, repData instances.RepetitionData) (*addrs.Reference, bool, tfdiags.Diagnostics) {

	// get the reference to lookup changes in the plan
	ref, diags := evalReplaceTriggeredByExpr(expr, repData)
	if diags.HasErrors() {
		return nil, false, diags
	}

	var changes []*plans.ResourceInstanceChangeSrc
	// store the address once we get it for validation
	var resourceAddr addrs.Resource

	// The reference is either a resource or resource instance
	switch sub := ref.Subject.(type) {
	case addrs.Resource:
		resourceAddr = sub
		rc := sub.Absolute(ctx.Path())
		changes = ctx.Changes().GetChangesForAbsResource(rc)
	case addrs.ResourceInstance:
		resourceAddr = sub.ContainingResource()
		rc := sub.Absolute(ctx.Path())
		change := ctx.Changes().GetResourceInstanceChange(rc, states.CurrentGen)
		if change != nil {
			// we'll generate an error below if there was no change
			changes = append(changes, change)
		}
	}

	// Do some validation to make sure we are expecting a change at all
	cfg := ctx.Evaluator.Config.Descendent(ctx.Path().Module())
	resCfg := cfg.Module.ResourceByAddr(resourceAddr)
	if resCfg == nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  `Reference to undeclared resource`,
			Detail:   fmt.Sprintf(`A resource %s has not been declared in %s`, ref.Subject, moduleDisplayAddr(ctx.Path())),
			Subject:  expr.Range().Ptr(),
		})
		return nil, false, diags
	}

	if len(changes) == 0 {
		// If the resource is valid there should always be at least one change.
		diags = diags.Append(fmt.Errorf("no change found for %s in %s", ref.Subject, moduleDisplayAddr(ctx.Path())))
		return nil, false, diags
	}

	// If we don't have a traversal beyond the resource, then we can just look
	// for any change.
	if len(ref.Remaining) == 0 {
		for _, c := range changes {
			switch c.ChangeSrc.Action {
			// Only immediate changes to the resource will trigger replacement.
			case plans.Update, plans.DeleteThenCreate, plans.CreateThenDelete:
				return ref, true, diags
			}
		}

		// no change triggered
		return nil, false, diags
	}

	// This must be an instances to have a remaining traversal, which means a
	// single change.
	change := changes[0]

	// Make sure the change is actionable. A create or delete action will have
	// a change in value, but are not valid for our purposes here.
	switch change.ChangeSrc.Action {
	case plans.Update, plans.DeleteThenCreate, plans.CreateThenDelete:
		// OK
	default:
		return nil, false, diags
	}

	// Since we have a traversal after the resource reference, we will need to
	// decode the changes, which means we need a schema.
	providerAddr := change.ProviderAddr
	schema, err := ctx.ProviderSchema(providerAddr)
	if err != nil {
		diags = diags.Append(err)
		return nil, false, diags
	}

	resAddr := change.Addr.ContainingResource().Resource
	resSchema, _ := schema.SchemaForResourceType(resAddr.Mode, resAddr.Type)
	ty := resSchema.ImpliedType()

	before, err := change.ChangeSrc.Before.Decode(ty)
	if err != nil {
		diags = diags.Append(err)
		return nil, false, diags
	}

	after, err := change.ChangeSrc.After.Decode(ty)
	if err != nil {
		diags = diags.Append(err)
		return nil, false, diags
	}

	path := traversalToPath(ref.Remaining)
	attrBefore, _ := path.Apply(before)
	attrAfter, _ := path.Apply(after)

	if attrBefore == cty.NilVal || attrAfter == cty.NilVal {
		replace := attrBefore != attrAfter
		return ref, replace, diags
	}

	replace := !attrBefore.RawEquals(attrAfter)

	return ref, replace, diags
}

func (ctx *BuiltinEvalContext) EvaluationScope(self addrs.Referenceable, source addrs.Referenceable, keyData InstanceKeyEvalData) *lang.Scope {
	if !ctx.pathSet {
		panic("context path not set")
	}
	data := &evaluationStateData{
		Evaluator:       ctx.Evaluator,
		ModulePath:      ctx.PathValue,
		InstanceKeyData: keyData,
		Operation:       ctx.Evaluator.Operation,
	}

	// ctx.PathValue is the path of the module that contains whatever
	// expression the caller will be trying to evaluate, so this will
	// activate only the experiments from that particular module, to
	// be consistent with how experiment checking in the "configs"
	// package itself works. The nil check here is for robustness in
	// incompletely-mocked testing situations; mc should never be nil in
	// real situations.
	mc := ctx.Evaluator.Config.DescendentForInstance(ctx.PathValue)

	if mc == nil || mc.Module.ProviderRequirements == nil {
		return ctx.Evaluator.Scope(data, self, source, nil)
	}

	ctx.FunctionLock.Lock()
	defer ctx.FunctionLock.Unlock()
	if ctx.FunctionCache == nil {
		names := make(map[string]addrs.Provider)

		// Providers must exist within required_providers to register their functions
		for name, provider := range mc.Module.ProviderRequirements.RequiredProviders {
			// Functions are only registered under their name, not their type name
			names[name] = provider.Type
		}

		ctx.FunctionCache = ctx.Plugins.Functions(names)
	}
	scope := ctx.Evaluator.Scope(data, self, source, ctx.FunctionCache)
	scope.SetActiveExperiments(mc.Module.ActiveExperiments)

	return scope
}

func (ctx *BuiltinEvalContext) Path() addrs.ModuleInstance {
	if !ctx.pathSet {
		panic("context path not set")
	}
	return ctx.PathValue
}

func (ctx *BuiltinEvalContext) SetRootModuleArgument(addr addrs.InputVariable, v cty.Value) {
	ctx.VariableValuesLock.Lock()
	defer ctx.VariableValuesLock.Unlock()

	log.Printf("[TRACE] BuiltinEvalContext: Storing final value for variable %s", addr.Absolute(addrs.RootModuleInstance))
	key := addrs.RootModuleInstance.String()
	args := ctx.VariableValues[key]
	if args == nil {
		args = make(map[string]cty.Value)
		ctx.VariableValues[key] = args
	}
	args[addr.Name] = v
}

func (ctx *BuiltinEvalContext) SetModuleCallArgument(callAddr addrs.ModuleCallInstance, varAddr addrs.InputVariable, v cty.Value) {
	ctx.VariableValuesLock.Lock()
	defer ctx.VariableValuesLock.Unlock()

	if !ctx.pathSet {
		panic("context path not set")
	}

	childPath := callAddr.ModuleInstance(ctx.PathValue)
	log.Printf("[TRACE] BuiltinEvalContext: Storing final value for variable %s", varAddr.Absolute(childPath))
	key := childPath.String()
	args := ctx.VariableValues[key]
	if args == nil {
		args = make(map[string]cty.Value)
		ctx.VariableValues[key] = args
	}
	args[varAddr.Name] = v
}

func (ctx *BuiltinEvalContext) GetVariableValue(addr addrs.AbsInputVariableInstance) cty.Value {
	ctx.VariableValuesLock.Lock()
	defer ctx.VariableValuesLock.Unlock()

	modKey := addr.Module.String()
	modVars := ctx.VariableValues[modKey]
	val, ok := modVars[addr.Variable.Name]
	if !ok {
		return cty.DynamicVal
	}
	return val
}

func (ctx *BuiltinEvalContext) Changes() *plans.ChangesSync {
	return ctx.ChangesValue
}

func (ctx *BuiltinEvalContext) State() *states.SyncState {
	return ctx.StateValue
}

func (ctx *BuiltinEvalContext) Checks() *checks.State {
	return ctx.ChecksValue
}

func (ctx *BuiltinEvalContext) RefreshState() *states.SyncState {
	return ctx.RefreshStateValue
}

func (ctx *BuiltinEvalContext) PrevRunState() *states.SyncState {
	return ctx.PrevRunStateValue
}

func (ctx *BuiltinEvalContext) InstanceExpander() *instances.Expander {
	return ctx.InstanceExpanderValue
}

func (ctx *BuiltinEvalContext) MoveResults() refactoring.MoveResults {
	return ctx.MoveResultsValue
}

func (ctx *BuiltinEvalContext) ImportResolver() *ImportResolver {
	return ctx.ImportResolverValue
}

func (ctx *BuiltinEvalContext) GetEncryption() encryption.Encryption {
	return ctx.Encryption
}
