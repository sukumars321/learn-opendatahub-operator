# Day 3 Quick Reference: Controllers & Reconciliation

## Core Controller Concepts

### The Reconciliation Loop
```
┌─────────────┐    Watch     ┌──────────────┐    Compare    ┌─────────────┐
│  Resource   │─────────────▶│  Controller  │──────────────▶│   Action    │
│   Events    │              │              │               │  Execution  │
└─────────────┘              └──────────────┘               └─────────────┘
                                     ▲                               │
                                     │          Status               │
                                     └───────────Update──────────────┘
```

### Controller Components
- **Watcher**: Monitors resource changes
- **Work Queue**: Buffers reconciliation requests
- **Reconciler**: Implements business logic
- **Status Reporter**: Updates resource status

## ODH Controller Architecture

### Hierarchy
```
DataScienceCluster Controller
├── DSCInitialization Controller
├── Dashboard Controller
├── Workbenches Controller
├── ModelMesh Controller
├── DataSciencePipelines Controller
└── ... (20+ component controllers)
```

### Key Files in ODH Codebase
```
controllers/
├── datasciencecluster/
│   └── datasciencecluster_controller.go    # Main orchestrator
├── components/
│   ├── dashboard.go                        # Dashboard component
│   ├── workbenches.go                     # Workbenches component
│   └── ...
└── dscinitialization/
    └── dscinitialization_controller.go    # Initialization controller
```

## Reconciliation Patterns

### Basic Reconcile Function Structure
```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the resource
    resource := &MyResource{}
    err := r.Get(ctx, req.NamespacedName, resource)
    if err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Handle deletion
    if resource.DeletionTimestamp != nil {
        return r.handleDeletion(ctx, resource)
    }

    // 3. Ensure finalizer
    if !controllerutil.ContainsFinalizer(resource, myFinalizer) {
        controllerutil.AddFinalizer(resource, myFinalizer)
        return ctrl.Result{}, r.Update(ctx, resource)
    }

    // 4. Reconcile desired state
    if err := r.reconcileComponents(ctx, resource); err != nil {
        return ctrl.Result{RequeueAfter: time.Minute}, err
    }

    // 5. Update status
    r.updateStatus(ctx, resource)

    return ctrl.Result{}, nil
}
```

### Component Management Pattern
```go
type ComponentInterface interface {
    ReconcileComponent(ctx context.Context, dsc *DataScienceCluster, platform Platform) error
    Cleanup(ctx context.Context, dsc *DataScienceCluster, platform Platform) error
    GetName() string
}

func (c *Component) ReconcileComponent(ctx context.Context, dsc *DataScienceCluster, platform Platform) error {
    switch dsc.Spec.Components.MyComponent.ManagementState {
    case "Managed":
        return c.deploy(ctx, dsc, platform)
    case "Removed":
        return c.Cleanup(ctx, dsc, platform)
    default:
        return nil
    }
}
```

## Status Management

### Status Structure
```yaml
status:
  phase: Ready|NotReady|Unknown
  conditions:
  - type: Ready
    status: "True|False|Unknown"
    reason: ReconcileCompleted
    message: "Human readable message"
    lastTransitionTime: "2024-09-30T10:30:00Z"
  - type: ComponentReady
    status: "True"
    reason: DeploymentAvailable
    message: "Component deployed successfully"
```

### Status Update Pattern
```go
func (r *Reconciler) updateStatus(ctx context.Context, resource *MyResource) {
    // Update individual conditions
    meta.SetStatusCondition(&resource.Status.Conditions, metav1.Condition{
        Type:    "Ready",
        Status:  metav1.ConditionTrue,
        Reason:  "ReconcileCompleted",
        Message: "All components ready",
    })

    // Update overall phase
    if allComponentsReady(resource.Status.Conditions) {
        resource.Status.Phase = "Ready"
    } else {
        resource.Status.Phase = "NotReady"
    }

    // Persist status
    r.Status().Update(ctx, resource)
}
```

## Watch Configuration

### Setting up Watches
```go
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&MyResource{}).                    // Primary resource
        Owns(&appsv1.Deployment{}).           // Owned child resources
        Owns(&corev1.Service{}).
        Watches(&source.Kind{Type: &corev1.ConfigMap{}},  // Related resources
               handler.EnqueueRequestsFromMapFunc(r.findOwners)).
        WithOptions(controller.Options{
            MaxConcurrentReconciles: 1,        // Serialize reconciliations
        }).
        Complete(r)
}
```

### Owner References
```yaml
metadata:
  ownerReferences:
  - apiVersion: datasciencecluster.opendatahub.io/v1
    kind: DataScienceCluster
    name: default-dsc
    uid: abc-123-def
    controller: true
    blockOwnerDeletion: true
```

## Useful Commands

### Monitoring Controllers
```bash
# Watch controller logs
oc logs -f deployment/opendatahub-operator-controller-manager -n opendatahub-operator-system

# Monitor specific resource types
oc get datasciencecluster,deployments,services -o wide -w

# Check status conditions
oc get datasciencecluster default-dsc -o jsonpath='{.status.conditions}' | jq

# Watch events
oc get events --watch --field-selector involvedObject.kind=DataScienceCluster
```

### Debugging Controllers
```bash
# Check controller metrics
oc port-forward svc/opendatahub-operator-controller-manager-metrics-service 8080 -n opendatahub-operator-system
curl http://localhost:8080/metrics | grep controller_runtime

# Examine owner references
oc get deployment odh-dashboard -o yaml | yq '.metadata.ownerReferences'

# Check finalizers
oc get datasciencecluster default-dsc -o jsonpath='{.metadata.finalizers}'
```

### Testing Controller Behavior
```bash
# Trigger reconciliation by changing spec
oc patch datasciencecluster default-dsc --type='merge' \
  -p='{"spec":{"components":{"dashboard":{"managementState":"Removed"}}}}'

# Simulate drift by deleting owned resource
oc delete deployment odh-dashboard

# Check reconciliation timing
oc describe datasciencecluster default-dsc | grep -A 20 "Status:"
```

## Common Reconciliation Results

```go
// Success, no requeue needed
return ctrl.Result{}, nil

// Error occurred, automatic retry with exponential backoff
return ctrl.Result{}, fmt.Errorf("failed to create deployment: %w", err)

// Explicit requeue after delay
return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil

// Immediate requeue (use sparingly)
return ctrl.Result{Requeue: true}, nil
```

## Error Handling Patterns

### Transient Errors
```go
if isTransientError(err) {
    log.Info("Transient error, will retry", "error", err)
    return ctrl.Result{RequeueAfter: time.Minute}, nil
}
```

### Permanent Errors
```go
if isPermanentError(err) {
    log.Error(err, "Permanent error, manual intervention needed")
    r.updateStatusWithError(ctx, resource, err)
    return ctrl.Result{}, nil // Don't requeue
}
```

### Resource Not Found
```go
err := r.Get(ctx, req.NamespacedName, resource)
if err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
}
```

## Performance Considerations

### Controller Efficiency
- Use **informers** (cached clients) for reads
- **Batch operations** when possible
- **Limit reconciliation frequency** with RequeueAfter
- **Use predicates** to filter unnecessary events
- **Implement proper indexing** for cross-resource lookups

### Memory Management
- **Don't cache large objects** unnecessarily
- **Use context cancellation** for long operations
- **Clean up resources** in finalizers
- **Monitor memory usage** with controller metrics

## Key Takeaways

1. **Controllers implement behavior** for CRDs through reconciliation loops
2. **Event-driven architecture** is more efficient than polling
3. **Owner references** enable automatic garbage collection
4. **Status conditions** provide rich feedback to users
5. **Finalizers** ensure orderly cleanup
6. **ODH uses hierarchical controllers** to manage complexity
7. **Reconciliation should be idempotent** and error-tolerant