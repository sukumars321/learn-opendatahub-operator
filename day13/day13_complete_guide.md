# Day 13: Watching and Event Filtering - Complete Study Guide

## Study Time: 40 minutes

## Learning Objectives
- Master the controller-runtime watch architecture and event flow mechanisms
- Analyze real ODH operator watch configurations and filtering strategies
- Understand event filtering with predicates and cross-resource watching patterns
- Implement efficient watch configurations for complex operator scenarios

---

## Part 1: Watch Architecture Fundamentals (10 minutes)

### The Event-Driven Controller Model

Kubernetes controllers are fundamentally event-driven systems. Instead of continuously polling the API server, controllers use the **watch** mechanism to efficiently receive notifications about resource changes.

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  API Server     │───▶│  Watch Stream    │───▶│  Event Queue    │
│  (Resource      │    │  (Filtered)      │    │  (Work Queue)   │
│   Changes)      │    └──────────────────┘    └─────────────────┘
└─────────────────┘             │                        │
                                 ▼                        ▼
                        ┌──────────────────┐    ┌─────────────────┐
                        │  Event Filters   │    │  Reconciler     │
                        │  (Predicates)    │    │  Execution      │
                        └──────────────────┘    └─────────────────┘
```

### Controller-Runtime Watch Mechanics

The controller-runtime framework provides several methods for setting up watches:

1. **`.For()`** - Watch the primary resource type the controller manages
2. **`.Owns()`** - Watch resources owned/created by the primary resource
3. **`.Watches()`** - Watch arbitrary resource types with custom logic

```go
// Basic controller setup showing different watch types
func (r *MyReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&myv1.MyResource{}).                    // Primary resource
        Owns(&appsv1.Deployment{}).                 // Owned resources
        Watches(&corev1.ConfigMap{},                // External resources
            reconciler.WithEventHandler(&handler.EnqueueRequestForOwner{
                OwnerType: &myv1.MyResource{},
                IsController: false,
            })).
        Complete(r)
}
```

### Event Flow and Work Queue

1. **Resource Change**: Someone creates, updates, or deletes a Kubernetes resource
2. **API Server Event**: The API server generates an event for the change
3. **Watch Stream**: Controllers receive the event through their watch streams
4. **Event Filtering**: Predicates filter events to determine relevance
5. **Work Queue**: Relevant events are queued for reconciliation
6. **Reconciler Execution**: The reconciler processes the queued work

---

## Part 2: ODH Operator Watch Patterns Analysis (15 minutes)

### DataScienceCluster Controller Watch Configuration

Let's examine how the ODH DataScienceCluster controller sets up its watches:

```go
// From internal/controller/datasciencecluster/datasciencecluster_controller.go
func (r *DataScienceClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Create a predicate for watching status changes on dependent resources
    componentsPredicate := dependent.New(dependent.WithWatchStatus(true))

    return ctrl.NewControllerManagedBy(mgr).
        For(&dscv1.DataScienceCluster{}).

        // Watch owned component resources with status monitoring
        Owns(&componentApi.Dashboard{}, reconciler.WithPredicates(componentsPredicate)).
        Owns(&componentApi.Workbenches{}, reconciler.WithPredicates(componentsPredicate)).
        Owns(&componentApi.DataSciencePipelines{}, reconciler.WithPredicates(componentsPredicate)).

        // Watch DSCInitialization for cross-resource coordination
        Watches(&dsciv2.DSCInitialization{},
            reconciler.WithEventMapper(func(ctx context.Context, _ client.Object) []reconcile.Request {
                return watchDataScienceClusters(ctx, mgr.GetClient())
            })).

        Complete(r)
}
```

### Component Controller Watch Strategies

ODH component controllers use sophisticated watch patterns. Here's the Workbenches controller:

```go
// From internal/controller/components/workbenches/workbenches_controller.go
func (r *WorkbenchesReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&componentApi.Workbenches{}).

        // Watch deployments with custom predicates
        Owns(&appsv1.Deployment{},
            reconciler.WithPredicates(resources.NewDeploymentPredicate())).

        // Watch CRDs for component availability
        Watches(&extv1.CustomResourceDefinition{},
            reconciler.WithEventHandler(handlers.ToNamed(componentApi.WorkbenchesInstanceName)),
            reconciler.WithPredicates(
                component.ForLabel(labels.ODH.Component(LegacyComponentName), labels.True))).

        Complete(r)
}
```

### Key ODH Watch Patterns

#### 1. Status-Aware Watching
ODH uses the `dependent` predicate to watch for status changes on component resources:

```go
// Custom predicate that watches status changes
componentsPredicate := dependent.New(dependent.WithWatchStatus(true))
```

This ensures the DataScienceCluster reconciler runs when component statuses change, allowing it to update the overall cluster status.

#### 2. Cross-Resource Event Mapping
ODH uses event mappers to trigger reconciliation across related resources:

```go
reconciler.WithEventMapper(func(ctx context.Context, _ client.Object) []reconcile.Request {
    return watchDataScienceClusters(ctx, mgr.GetClient())
})
```

When a DSCInitialization changes, this mapper finds all related DataScienceClusters and queues them for reconciliation.

#### 3. Label-Based Filtering
Component controllers use label selectors to watch only relevant CRDs:

```go
reconciler.WithPredicates(
    component.ForLabel(labels.ODH.Component(LegacyComponentName), labels.True))
```

---

## Part 3: Event Filtering with Predicates (10 minutes)

### Understanding Predicates

Predicates are functions that filter events before they reach the reconciler. They implement the `predicate.Predicate` interface:

```go
type Predicate interface {
    Create(event.CreateEvent) bool
    Delete(event.DeleteEvent) bool
    Update(event.UpdateEvent) bool
    Generic(event.GenericEvent) bool
}
```

### ODH Custom Predicates

#### 1. Dependent Predicate for Status Watching

```go
// From pkg/controller/predicates/dependent/predicate.go
type Dependent struct {
    watchStatus bool
}

func (p *Dependent) Update(e event.UpdateEvent) bool {
    if !p.watchStatus {
        return false
    }

    // Only trigger reconciliation if status changed
    oldStatus := getResourceStatus(e.ObjectOld)
    newStatus := getResourceStatus(e.ObjectNew)

    return !reflect.DeepEqual(oldStatus, newStatus)
}
```

This predicate ensures reconciliation only happens when status fields change, not on every spec update.

#### 2. Deployment Predicate for Resource Watching

```go
// From pkg/controller/predicates/resources/deployment_predicate.go
func NewDeploymentPredicate() predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            oldDeployment := e.ObjectOld.(*appsv1.Deployment)
            newDeployment := e.ObjectNew.(*appsv1.Deployment)

            // Only reconcile on meaningful deployment changes
            return deploymentConditionsChanged(oldDeployment, newDeployment) ||
                   deploymentReplicasChanged(oldDeployment, newDeployment)
        },
    }
}
```

### Predicate Best Practices

1. **Filter Early**: Use predicates to reduce unnecessary reconciliation
2. **Status vs Spec**: Separate predicates for status and spec changes
3. **Performance**: Keep predicate logic fast and efficient
4. **Debugging**: Log predicate decisions for troubleshooting

```go
// Example of a well-designed predicate
func NewComponentPredicate(logger logr.Logger) predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            old := e.ObjectOld.(*componentApi.Component)
            new := e.ObjectNew.(*componentApi.Component)

            // Check if management state changed
            if old.Spec.ManagementState != new.Spec.ManagementState {
                logger.Info("Component management state changed",
                    "old", old.Spec.ManagementState,
                    "new", new.Spec.ManagementState)
                return true
            }

            // Check if conditions changed
            return !reflect.DeepEqual(old.Status.Conditions, new.Status.Conditions)
        },
    }
}
```

---

## Part 4: Cross-Resource Watching Strategies (5 minutes)

### Event Mappers for Complex Relationships

When resources have complex relationships, event mappers help trigger appropriate reconciliation:

```go
// Example: When a ConfigMap changes, reconcile all components using it
func mapConfigMapToComponents(ctx context.Context, client client.Client) handler.MapFunc {
    return func(ctx context.Context, obj client.Object) []reconcile.Request {
        configMap := obj.(*corev1.ConfigMap)

        // Find all components that reference this ConfigMap
        var components componentApi.ComponentList
        if err := client.List(ctx, &components); err != nil {
            return []reconcile.Request{}
        }

        var requests []reconcile.Request
        for _, component := range components.Items {
            if referencesConfigMap(component, configMap) {
                requests = append(requests, reconcile.Request{
                    NamespacedName: types.NamespacedName{
                        Name:      component.Name,
                        Namespace: component.Namespace,
                    },
                })
            }
        }
        return requests
    }
}
```

### Watch Hierarchies in ODH

ODH implements a hierarchical watch strategy:

1. **DSCInitialization** → Triggers DataScienceCluster reconciliation
2. **DataScienceCluster** → Watches component status changes
3. **Components** → Watch their owned resources (Deployments, Services, etc.)

This creates an efficient event propagation system where changes bubble up through the resource hierarchy.

---

## Key Takeaways

### Watch Design Principles
1. **Efficiency First**: Use predicates to filter unnecessary events
2. **Clear Relationships**: Map resource relationships with event mappers
3. **Status Awareness**: Separate watching of spec vs status changes
4. **Debugging Support**: Add logging to watch configurations

### Common Patterns
- **Owned Resource Watching**: Use `.Owns()` with predicates for child resources
- **Cross-Resource Coordination**: Use `.Watches()` with event mappers
- **Label-Based Filtering**: Filter watches using label selectors in predicates
- **Status Monitoring**: Use status-aware predicates for state propagation

### Performance Considerations
- Predicates run for every event - keep them fast
- Event mappers can generate multiple reconciliation requests
- Watch too many resources and you'll overload the reconciler
- Watch too few and you'll miss important changes

Understanding these watch patterns is crucial for building efficient, responsive operators that react appropriately to cluster state changes while maintaining optimal performance.