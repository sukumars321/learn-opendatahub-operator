# Day 6: Controller-Runtime Framework - Complete Study Guide

*Study Time: 45 minutes*

## Introduction (5 minutes)

Controller-runtime is a high-level framework that sits on top of client-go, making operator development much easier. It provides:
- **Manager pattern** for organizing multiple controllers
- **Builder pattern** for configuring controllers declaratively
- **Reconciliation abstractions** that handle the complexity of event processing
- **Testing utilities** for unit and integration testing

**Key Concept**: While client-go gives you the building blocks, controller-runtime gives you the architecture patterns that most operators need.

## Controller-Runtime vs Client-go (8 minutes)

### The Abstraction Ladder

```
┌─────────────────────────────────────┐
│ Your Business Logic (Reconcile)     │  ← You write this
├─────────────────────────────────────┤
│ Controller-Runtime Framework        │  ← Handles complexity
├─────────────────────────────────────┤
│ Client-go Library                   │  ← Low-level primitives
├─────────────────────────────────────┤
│ Kubernetes API Server               │  ← The truth
└─────────────────────────────────────┘
```

### What Controller-Runtime Provides

**Without controller-runtime** (pure client-go):
```go
// You'd need to manage all this complexity
informer := factory.ForResource(gvr).Informer()
informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
        workqueue.Add(...)
    },
    UpdateFunc: func(oldObj, newObj interface{}) {
        workqueue.Add(...)
    },
    DeleteFunc: func(obj interface{}) {
        workqueue.Add(...)
    },
})
```

**With controller-runtime**:
```go
// Much simpler declarative setup
ctrl.NewControllerManagedBy(mgr).
    For(&myapi.MyResource{}).
    Owns(&appsv1.Deployment{}).
    Complete(r)
```

### Benefits of Controller-Runtime

1. **Declarative Configuration**: Describe what you want, not how to achieve it
2. **Consistent Patterns**: All operators follow similar structures
3. **Built-in Best Practices**: Event filtering, work queues, error handling
4. **Testing Support**: Utilities for unit and integration testing
5. **Webhook Integration**: Easy admission controller setup

## Manager Pattern (10 minutes)

### What is a Manager?

The Manager is the central coordinator that:
- **Manages the shared cache** for all controllers
- **Coordinates controller lifecycle** (start, stop, health)
- **Provides shared clients** for Kubernetes API access
- **Handles leader election** for high availability
- **Manages webhooks** and other components

### ODH Manager Setup

**File Reference**: `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/main.go`

```go
func main() {
    // Create the manager
    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:                 scheme,
        MetricsBindAddress:     metricsAddr,
        Port:                   9443,
        HealthProbeBindAddress: probeAddr,
        LeaderElection:         enableLeaderElection,
        LeaderElectionID:       "opendatahub-operator",
    })

    // Add controllers to the manager
    if err = (&controllers.DataScienceClusterReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create controller", "controller", "DataScienceCluster")
        os.Exit(1)
    }

    // Start the manager
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        setupLog.Error(err, "problem running manager")
        os.Exit(1)
    }
}
```

### Manager Responsibilities

1. **Shared Resources**:
   - Single Kubernetes client shared by all controllers
   - Shared informer cache to reduce API server load
   - Common metrics and health endpoints

2. **Lifecycle Management**:
   - Start all controllers simultaneously
   - Graceful shutdown coordination
   - Signal handling for process termination

3. **Leader Election**:
   - Only one instance actively reconciling (in multi-replica deployments)
   - Automatic failover between replicas
   - Prevents split-brain scenarios

## Builder Pattern for Controllers (12 minutes)

### Declarative Controller Setup

Controller-runtime uses the builder pattern to make controller setup declarative and readable.

### Basic Builder Example

```go
err := ctrl.NewControllerManagedBy(mgr).
    For(&dsci.DataScienceCluster{}).           // Primary resource to watch
    Owns(&corev1.ConfigMap{}).                 // Owned resources to watch
    Owns(&appsv1.Deployment{}).                // Multiple owned types
    WithOptions(controller.Options{
        MaxConcurrentReconciles: 1,            // Concurrency control
    }).
    WithEventFilter(predicate.GenerationChangedPredicate{}). // Event filtering
    Complete(r)                                // Build and register
```

### Understanding Builder Methods

**`.For(resource)`** - Primary Resource
- Creates informer for the main resource type
- Reconcile is triggered when this resource changes
- Usually your custom resource (CRD)

**`.Owns(resource)`** - Owned Resources
- Creates informers for resources created by your controller
- Reconcile is triggered when owned resources change
- Establishes parent-child relationship

**`.Watches(source, handler)`** - Custom Watching
- Watch arbitrary resources with custom handlers
- More flexible than For/Owns pattern
- Useful for external dependencies

### ODH Controller Setup Analysis

**File Reference**: Look for `SetupWithManager` functions in ODH controllers:

```go
func (r *DataScienceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&dsci.DataScienceCluster{}).
        Owns(&corev1.Secret{}).
        Owns(&corev1.ConfigMap{}).
        Owns(&rbacv1.ClusterRole{}).
        Owns(&rbacv1.ClusterRoleBinding{}).
        WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
        Complete(r)
}
```

**Pattern Analysis**:
- **Primary Resource**: DataScienceCluster (the main CRD)
- **Owned Resources**: Secret, ConfigMap, RBAC resources
- **Concurrency**: Limited to 1 for safety
- **Event Filtering**: Uses default generation-based filtering

### Event Filtering

```go
// Built-in predicates
WithEventFilter(predicate.GenerationChangedPredicate{})    // Only spec changes
WithEventFilter(predicate.NamespaceChangedPredicate{})     // Namespace events
WithEventFilter(predicate.LabelChangedPredicate{})         // Label changes

// Custom predicates
WithEventFilter(predicate.Funcs{
    UpdateFunc: func(e event.UpdateEvent) bool {
        // Custom logic for when to reconcile
        return shouldReconcile(e.ObjectOld, e.ObjectNew)
    },
})
```

## Reconciliation Flow (10 minutes)

### The Reconcile Function

The heart of any controller is the Reconcile function:

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the resource
    // 2. Determine desired state
    // 3. Compare with current state
    // 4. Take action to align them
    // 5. Update status
    // 6. Return result
}
```

### Reconcile Input and Output

**Input**: `ctrl.Request`
```go
type Request struct {
    NamespacedName types.NamespacedName  // Which resource to reconcile
}
```

**Output**: `(ctrl.Result, error)`
```go
type Result struct {
    Requeue      bool          // Requeue immediately
    RequeueAfter time.Duration // Requeue after delay
}
```

### Reconcile Return Patterns

```go
// Success - don't requeue
return ctrl.Result{}, nil

// Requeue immediately (use sparingly)
return ctrl.Result{Requeue: true}, nil

// Requeue after delay (common for polling)
return ctrl.Result{RequeueAfter: time.Minute * 5}, nil

// Error - requeue with exponential backoff
return ctrl.Result{}, fmt.Errorf("failed to create deployment: %w", err)

// Temporary error - requeue after delay
if isTemporaryError(err) {
    return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}
return ctrl.Result{}, err
```

### ODH Reconcile Pattern

**File Reference**: Examine ODH controller reconcile functions:

```go
func (r *DataScienceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("datasciencecluster", req.NamespacedName)

    // 1. Fetch the DataScienceCluster instance
    var instance dsci.DataScienceCluster
    if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Handle deletion (finalizers)
    if instance.DeletionTimestamp != nil {
        return r.reconcileDelete(ctx, &instance)
    }

    // 3. Handle creation/update
    return r.reconcileNormal(ctx, &instance)
}
```

**Common Pattern Elements**:
1. **Logging with context**: Include resource identifier
2. **Fetch resource**: Handle not found gracefully
3. **Deletion handling**: Check DeletionTimestamp
4. **Separate reconcile logic**: Different functions for create/update vs delete

## Advanced Controller-Runtime Features (5 minutes)

### Client Improvements over Raw Client-go

**Unified Client Interface**:
```go
// Single client for all operations
err := r.Get(ctx, key, &obj)           // Works with any resource
err := r.Create(ctx, &obj)             // Type-safe operations
err := r.Update(ctx, &obj)             // Automatic conflict handling
err := r.Status().Update(ctx, &obj)    // Separate status subresource
```

**Automatic Retries**:
- Conflict errors are automatically retried
- Stale cache reads trigger refresh and retry
- Network errors have built-in backoff

### Indexing for Efficient Queries

```go
// Set up indexing for efficient lookups
mgr.GetFieldIndexer().IndexField(ctx, &appsv1.Deployment{}, "spec.selector.matchLabels", func(obj client.Object) []string {
    deployment := obj.(*appsv1.Deployment)
    return []string{deployment.Spec.Selector.MatchLabels["app"]}
})

// Query using the index
var deployments appsv1.DeploymentList
err := r.List(ctx, &deployments, client.MatchingFields{"spec.selector.matchLabels": "myapp"})
```

### Metrics and Observability

```go
// Built-in metrics
// - controller_runtime_reconcile_total
// - controller_runtime_reconcile_time_seconds
// - controller_runtime_reconcile_errors_total

// Custom metrics
reconcileCounter := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "my_controller_reconciles_total",
    },
    []string{"result"},
)
```

## Testing with Controller-Runtime (5 minutes)

### EnvTest Integration

Controller-runtime provides excellent testing support:

```go
var testEnv *envtest.Environment
var cfg *rest.Config
var k8sClient client.Client

BeforeSuite(func() {
    testEnv = &envtest.Environment{
        CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
    }

    cfg, err = testEnv.Start()
    Expect(err).NotTo(HaveOccurred())

    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    Expect(err).NotTo(HaveOccurred())
})
```

### Testing Patterns

1. **Unit Testing**: Test reconcile logic with fake clients
2. **Integration Testing**: Test with real Kubernetes API (envtest)
3. **End-to-End Testing**: Test complete operator deployment

## Summary and Key Takeaways

### What You've Learned

1. **Controller-Runtime Architecture**: Manager, Builder, Reconciler patterns
2. **Abstraction Benefits**: How it simplifies client-go complexity
3. **ODH Usage Patterns**: Real-world controller setup and reconciliation
4. **Best Practices**: Event filtering, error handling, testing approaches

### Key Concepts to Remember

- **Manager coordinates everything** - shared resources, lifecycle, leader election
- **Builder pattern makes setup declarative** - describe what you want
- **Reconcile function is where your logic lives** - idempotent operations
- **Return values control requeue behavior** - success, error, or retry patterns

### Connection to Tomorrow

Day 7 will be a review day where you'll consolidate your understanding of the foundation concepts from this week: Kubernetes APIs, CRDs, Controllers, Go patterns, client-go, and controller-runtime.

**Time Check**: You should have spent about 45 minutes understanding these concepts. Focus on recognizing patterns rather than memorizing implementation details.