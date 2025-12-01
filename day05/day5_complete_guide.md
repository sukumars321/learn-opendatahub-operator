# Day 5: Client-go Library Basics - Complete Study Guide

*Study Time: 45 minutes*

## Introduction (5 minutes)

Client-go is the official Go client library for Kubernetes. It's the foundation that makes operator development possible by providing:
- Typed clients for Kubernetes resources
- Informers for efficient watching
- Work queues for reliable processing
- Discovery mechanisms for API exploration

**Key Concept**: While controller-runtime abstracts much of client-go's complexity, understanding client-go helps you debug issues and understand what's happening under the hood.

## Client-go Architecture Overview (10 minutes)

### Core Components

```
┌─────────────────┐
│   Your Operator │
└─────┬───────────┘
      │
┌─────▼───────────┐
│ Controller-Runtime│ ← Abstracts client-go complexity
└─────┬───────────┘
      │
┌─────▼───────────┐
│   Client-go     │ ← Low-level Kubernetes client
└─────┬───────────┘
      │
┌─────▼───────────┐
│ Kubernetes API  │
└─────────────────┘
```

### Client Types in Client-go

1. **Clientset** - Typed clients for built-in resources
2. **Dynamic Client** - Untyped client for any resource
3. **Discovery Client** - Explore available APIs
4. **RESTClient** - Low-level HTTP client

## Clientset: Typed Kubernetes Clients (10 minutes)

### What is a Clientset?

A clientset provides typed access to Kubernetes resources. Each API group has its own clientset.

```go
// Example: Core v1 resources
coreV1Client := clientset.CoreV1()
pods := coreV1Client.Pods("namespace")
deployments := clientset.AppsV1().Deployments("namespace")
```

### ODH Operator Clientset Usage

Let's examine how ODH uses clientsets:

**File Reference**: `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/controllers/datasciencecluster_controller.go`

```go
// ODH creates clients for different resource types
func (r *DataScienceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Controller-runtime provides a client that combines multiple clientsets
    var dsc dsci.DataScienceCluster
    if err := r.Get(ctx, req.NamespacedName, &dsc); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Behind the scenes, this uses client-go clientsets
}
```

**Key Pattern**: ODH uses controller-runtime's client, which internally manages clientsets for you.

### Built-in vs Custom Resources

```go
// Built-in resources (use standard clientsets)
clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})

// Custom resources (use dynamic client or generated clientsets)
// ODH generates its own clientsets for custom resources
```

## Dynamic Client: Flexible Resource Access (8 minutes)

### When to Use Dynamic Client

Dynamic client is useful for:
- Working with custom resources without generated clients
- Building generic tools that work with any resource
- Runtime resource discovery

```go
// Dynamic client example
dynamicClient := dynamic.NewForConfigOrDie(config)

// Create a resource reference
gvr := schema.GroupVersionResource{
    Group:    "datasciencecluster.opendatahub.io",
    Version:  "v1",
    Resource: "datascienceclusters",
}

// Get a resource
obj, err := dynamicClient.Resource(gvr).
    Namespace("namespace").
    Get(ctx, "name", metav1.GetOptions{})
```

### ODH's Approach

ODH primarily uses controller-runtime's client, which handles both typed and dynamic access:

```go
// Controller-runtime client handles both cases seamlessly
var configMap corev1.ConfigMap
err := r.Get(ctx, types.NamespacedName{Name: "config", Namespace: "ns"}, &configMap)

var dsc dsci.DataScienceCluster
err := r.Get(ctx, req.NamespacedName, &dsc)
```

## Informers: Efficient Resource Watching (12 minutes)

### What are Informers?

Informers provide efficient, cached access to Kubernetes resources:
- **Watch** API server for changes
- **Cache** resources locally
- **Notify** your code of changes
- **List** cached resources efficiently

### Informer Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Kubernetes API │────▶│     Informer    │────▶│   Your Code     │
└─────────────────┘     │                 │     └─────────────────┘
                        │ ┌─────────────┐ │
                        │ │    Cache    │ │
                        │ └─────────────┘ │
                        └─────────────────┘
```

### How ODH Uses Informers

**File Reference**: Look for `SetupWithManager` functions in ODH controllers:

```go
func (r *DataScienceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&dsci.DataScienceCluster{}).    // Creates informer for DSC
        Owns(&corev1.ConfigMap{}).          // Creates informer for ConfigMaps
        Owns(&appsv1.Deployment{}).         // Creates informer for Deployments
        Complete(r)
}
```

**Behind the scenes**: Controller-runtime creates informers for each resource type you specify.

### Informer Benefits

1. **Efficiency**: Local cache reduces API server load
2. **Performance**: No need to poll API server
3. **Consistency**: Guaranteed event ordering
4. **Reliability**: Handles connection issues automatically

### Event Types

Informers generate three types of events:
```go
// Add - new resource created
func (r *Controller) OnAdd(obj interface{}) {}

// Update - existing resource modified
func (r *Controller) OnUpdate(oldObj, newObj interface{}) {}

// Delete - resource removed
func (r *Controller) OnDelete(obj interface{}) {}
```

## Work Queues: Reliable Processing (5 minutes)

### Why Work Queues?

Work queues ensure reliable event processing:
- **Deduplication**: Multiple events for same resource are merged
- **Rate Limiting**: Prevents overwhelming your controller
- **Retry Logic**: Failed reconciliations are retried with backoff

```go
// Simplified work queue flow
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Event   │───▶│ Queue   │───▶│Process  │───▶│ Done    │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
                      ▲              │
                      │              ▼
                      └──── Retry ───┘
```

### ODH's Queue Usage

Controller-runtime manages work queues automatically:

```go
func (r *DataScienceClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // This function is called from a work queue
    // req.NamespacedName contains the resource to process

    // Return values control queue behavior:
    return ctrl.Result{}, nil           // Success - remove from queue
    return ctrl.Result{Requeue: true}, nil  // Requeue immediately
    return ctrl.Result{RequeueAfter: time.Minute}, nil  // Requeue after delay
    return ctrl.Result{}, err           // Error - requeue with backoff
}
```

## Client Configuration and Authentication (5 minutes)

### Getting Kubernetes Config

```go
// In-cluster configuration (when running in pod)
config, err := rest.InClusterConfig()

// Out-of-cluster configuration (development)
config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

// ODH operator uses controller-runtime's manager
mgr, err := ctrl.NewManager(cfg, ctrl.Options{
    Scheme: scheme,
    // ... other options
})
```

### Authentication Methods

1. **Service Account Token** (in-cluster)
2. **Kubeconfig File** (development)
3. **Token File** (CI/CD)
4. **OIDC** (enterprise)

ODH operator typically runs with a service account that has appropriate RBAC permissions.

## Common Client-go Patterns in ODH (5 minutes)

### Pattern 1: Resource CRUD Operations

```go
// Read
var dsc dsci.DataScienceCluster
err := r.Get(ctx, req.NamespacedName, &dsc)

// Create
configMap := &corev1.ConfigMap{...}
err := r.Create(ctx, configMap)

// Update
dsc.Status.Phase = "Ready"
err := r.Status().Update(ctx, &dsc)

// Delete (rarely used directly)
err := r.Delete(ctx, object)
```

### Pattern 2: Listing Resources

```go
// List all resources of a type
var configMaps corev1.ConfigMapList
err := r.List(ctx, &configMaps, client.InNamespace("namespace"))

// List with label selector
err := r.List(ctx, &configMaps,
    client.InNamespace("namespace"),
    client.MatchingLabels{"app": "odh"})
```

### Pattern 3: Owner References

```go
// Set owner reference (for garbage collection)
err := controllerutil.SetControllerReference(&dsc, configMap, r.Scheme)
```

## Summary and Key Takeaways

### What You've Learned

1. **Client-go Architecture**: Clientsets, dynamic clients, informers, and work queues
2. **ODH Integration**: How ODH uses controller-runtime to abstract client-go
3. **Efficient Patterns**: Informers for watching, queues for processing
4. **Common Operations**: CRUD operations and listing patterns

### Key Concepts to Remember

- **Informers are essential** for efficient Kubernetes programming
- **Controller-runtime abstracts complexity** while using client-go underneath
- **Work queues ensure reliability** in event processing
- **RBAC permissions** are required for client operations

### Connection to Tomorrow

Day 6 will explore controller-runtime, which builds on client-go concepts to provide higher-level abstractions that make operator development much easier.

**Time Check**: You should have spent about 45 minutes on this guide. If you're running over, focus on the key concepts and patterns rather than memorizing every detail.