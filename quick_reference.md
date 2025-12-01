# OpenDataHub Operator - Quick Reference Guide

A handy reference for key concepts, commands, and code locations while studying.

## Key Technologies Overview

### Core Technologies
- **Kubebuilder**: v4 - Operator development framework
- **Controller-Runtime**: High-level controller framework
- **Client-go**: Kubernetes API client library
- **OLM**: Operator Lifecycle Manager for packaging
- **Go**: 1.24.4 - Programming language

### Testing & Quality
- **Ginkgo/Gomega**: BDD testing framework
- **Envtest**: Kubernetes API server for testing
- **Scorecard**: OLM bundle testing
- **Prometheus**: Monitoring and alerting

### Deployment & Packaging
- **Kustomize**: Configuration management
- **Docker**: Multi-stage container builds
- **OpenShift**: Production deployment platform

## Important File Locations

### Core Configuration
```
PROJECT                              # Kubebuilder project configuration
go.mod                              # Go module dependencies
main.go                             # Operator entry point
Makefile                            # Build and development commands
```

### API Definitions
```
apis/
├── datasciencecluster/v1/          # Main cluster CRD
├── dscinitialization/v1/           # Initialization CRD
├── components/platform.opendatahub.io/v1alpha1/  # Component CRDs
└── services/platform.opendatahub.io/v1alpha1/    # Service CRDs
```

### Controllers
```
controllers/
├── datasciencecluster_controller.go      # Main DSC controller
├── dscinitialization_controller.go       # Initialization controller
└── components/                           # Component controllers
```

### Generated Code
```
config/
├── crd/bases/                      # Generated CRD manifests
├── rbac/                          # Generated RBAC rules
├── webhook/                       # Webhook configurations
└── manager/                       # Manager deployment
```

### OLM Bundle
```
bundle/
├── manifests/                     # OLM bundle manifests
├── metadata/                      # Bundle metadata
└── tests/scorecard/               # Scorecard test configuration
```

## Essential Commands

### Development
```bash
# Generate code and manifests
make generate

# Build operator binary
make build

# Run tests
make test

# Install CRDs
make install

# Run locally (outside cluster)
make run

# Build and push container image
make docker-build docker-push IMG=<registry>/opendatahub-operator:tag
```

### OLM Bundle
```bash
# Generate bundle
make bundle IMG=<registry>/opendatahub-operator:tag

# Build bundle image
make bundle-build BUNDLE_IMG=<registry>/opendatahub-operator-bundle:tag

# Validate bundle
operator-sdk bundle validate ./bundle

# Run scorecard tests
operator-sdk scorecard bundle
```

### Kubebuilder Shortcuts
```bash
# Create new API
kubebuilder create api --group <group> --version <version> --kind <Kind>

# Create webhook
kubebuilder create webhook --group <group> --version <version> --kind <Kind> --programmatic-validation
```

## Key Code Patterns

### Basic Controller Setup
```go
func (r *MyReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&myapi.MyKind{}).
        Owns(&appsv1.Deployment{}).
        WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
        WithEventFilter(predicate.GenerationChangedPredicate{}).
        Complete(r)
}
```

### Reconciler Structure
```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("mykind", req.NamespacedName)

    // Fetch the resource
    var instance myapi.MyKind
    if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Reconciliation logic here

    // Update status
    if err := r.Status().Update(ctx, &instance); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

### Condition Management
```go
// Setting conditions
meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
    Type:    "Ready",
    Status:  metav1.ConditionTrue,
    Reason:  "ReconcileSuccess",
    Message: "Resource reconciled successfully",
})
```

## Common Kubebuilder Markers

### RBAC Markers
```go
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=mygroup,resources=mykinds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup,resources=mykinds/status,verbs=get;update;patch
```

### CRD Markers
```go
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
```

### Validation Markers
```go
//+kubebuilder:validation:Required
//+kubebuilder:validation:Optional
//+kubebuilder:validation:Minimum=1
//+kubebuilder:validation:Maximum=100
//+kubebuilder:validation:Pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
```

### Webhook Markers
```go
//+kubebuilder:webhook:path=/mutate-mygroup-v1-mykind,mutating=true,failurePolicy=fail,groups=mygroup,resources=mykinds,verbs=create;update,versions=v1,name=mmykind.mygroup
//+kubebuilder:webhook:path=/validate-mygroup-v1-mykind,mutating=false,failurePolicy=fail,groups=mygroup,resources=mykinds,verbs=create;update,versions=v1,name=vmykind.mygroup
```

## ODH-Specific Patterns

### Component Interface
```go
type ComponentInterface interface {
    ReconcileComponent(cli client.Client, logger logr.Logger, owner metav1.Object, dscispec *dsciv1.DSCInitializationSpec) error
    Cleanup(cli client.Client, owner metav1.Object) error
    GetComponentName() string
}
```

### Action Pattern
```go
type Action interface {
    Execute(ctx context.Context) error
}

// Common actions: deploy, cleanup, update status
```

### Component Status Pattern
```go
type ComponentStatus struct {
    Phase      string             `json:"phase,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

## Troubleshooting Commands

### Debug Controller
```bash
# Check controller logs
kubectl logs -n opendatahub-operator-system deployment/opendatahub-operator-controller-manager

# Check CRD installation
kubectl get crd | grep opendatahub

# Check webhook configuration
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations

# Check RBAC
kubectl auth can-i create datascienceclusters --as=system:serviceaccount:opendatahub-operator-system:opendatahub-operator-controller-manager
```

### Debug Resources
```bash
# Check resource status
kubectl get datasciencecluster -o yaml
kubectl describe datasciencecluster <name>

# Check events
kubectl get events --sort-by='.metadata.creationTimestamp'

# Check finalizers
kubectl get <resource> -o jsonpath='{.metadata.finalizers}'
```

## Useful Environment Variables

```bash
# Development
export KUBEBUILDER_ASSETS=/usr/local/kubebuilder/bin
export USE_EXISTING_CLUSTER=true

# Testing
export GINKGO_FOCUS="MyController"
export GINKGO_SKIP="integration"

# OLM
export BUNDLE_IMG=registry.redhat.io/ubi8/ubi:latest
export CATALOG_IMG=my-catalog:latest
```

## Study Tips

### Reading Code Efficiently
1. Start with `main.go` to understand entry point
2. Look at type definitions in `apis/` directories
3. Follow controller setup in `controllers/`
4. Trace reconciliation logic step by step
5. Check tests for usage examples

### Understanding Flow
1. CRD defines the API
2. Controller watches for changes
3. Reconciler implements business logic
4. Status reflects current state
5. Events provide audit trail

### Common Gotchas
- Always check for nil pointers
- Handle resource not found errors gracefully
- Update status separately from spec
- Use owner references for garbage collection
- Test with realistic scenarios

## Next Steps After Study

### Build Your Own Operator
1. Choose a simple use case
2. Design your CRDs
3. Implement basic controller
4. Add webhooks for validation
5. Create OLM bundle
6. Add comprehensive tests

### Contribute to ODH
1. Set up development environment
2. Pick a good first issue
3. Follow contribution guidelines
4. Write tests for your changes
5. Submit pull request

### Advanced Topics
- Multi-cluster operators
- Operator SDK advanced features
- Custom schedulers
- Extended APIs
- Performance optimization