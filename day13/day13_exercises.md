# Day 13: Watching and Event Filtering - Offline Exercises

## Exercise Time: 20-25 minutes

These exercises help you practice watch pattern design and event filtering concepts without requiring access to a live ODH environment. They focus on understanding the theory and applying it to design scenarios.

---

## Exercise 1: Watch Configuration Design (8 minutes)

### Scenario: Multi-Tenant ML Platform Controller

You're designing a controller for a `MLPlatform` custom resource that manages machine learning workloads in a multi-tenant environment.

**Requirements:**
- Primary resource: `MLPlatform` (namespace-scoped)
- Manages: `Deployments`, `Services`, `ConfigMaps`, `Secrets`
- Must react to changes in `Nodes` (for resource allocation)
- Should watch `ResourceQuotas` (for quota enforcement)
- Needs to monitor `PersistentVolumeClaims` for storage

### Task 1.1: Basic Watch Setup (3 minutes)

Design the `SetupWithManager` method. For each watch type, specify:
- Whether to use `.For()`, `.Owns()`, or `.Watches()`
- The resource type being watched
- Justification for the choice

```go
func (r *MLPlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        // YOUR DESIGN HERE
        Complete(r)
}
```

**Questions to consider:**
- Which resources are owned vs external?
- What relationships exist between resources?
- Which watches might generate the most events?

### Task 1.2: Predicate Design (5 minutes)

Design predicates for these scenarios:

1. **Node Predicate**: Only reconcile when node capacity or availability changes
2. **ConfigMap Predicate**: Only reconcile when specific ConfigMaps change (those with label `ml-platform=config`)
3. **PVC Predicate**: Only reconcile when PVC status changes (not spec)

For each predicate, write pseudo-code explaining:
- Which event types (Create/Update/Delete) should trigger reconciliation
- What specific conditions should be checked
- How to optimize for performance

---

## Exercise 2: Event Filtering Analysis (7 minutes)

### Scenario: High-Traffic Environment

Your operator runs in a cluster with high resource churn. You're seeing performance issues due to excessive reconciliation.

**Current Problems:**
- Reconciler runs on every Pod status update
- ConfigMap changes trigger reconciliation even for unrelated data
- Node updates cause reconciliation for all MLPlatforms

### Task 2.1: Predicate Optimization (4 minutes)

For each problem, design an optimized predicate:

1. **Pod Status Problem**:
   ```go
   // Current: Reconciles on ANY pod change
   Owns(&corev1.Pod{})

   // Your optimized version with predicate:
   // Owns(&corev1.Pod{}, reconciler.WithPredicates(?))
   ```

2. **ConfigMap Problem**:
   ```go
   // Current: Watches ALL configmaps
   Watches(&corev1.ConfigMap{}, /* some handler */)

   // Your optimized version:
   // What predicate logic would filter irrelevant ConfigMaps?
   ```

3. **Node Problem**:
   ```go
   // Current: Reconciles on ANY node change
   Watches(&corev1.Node{}, /* handler */)

   // Your optimized version:
   // What node changes actually matter for ML workloads?
   ```

### Task 2.2: Performance Impact Analysis (3 minutes)

Calculate the theoretical impact of your optimizations:

**Assumptions:**
- 100 MLPlatform instances in cluster
- 1000 Pod status updates per minute
- 50 ConfigMap changes per minute
- 10 Node updates per minute

**Current Reconciliation Rate**: _____ per minute
**Optimized Reconciliation Rate**: _____ per minute
**Improvement Factor**: _____

---

## Exercise 3: Cross-Resource Watching Strategy (10 minutes)

### Scenario: Dependent Resource Management

Design a watching strategy for a complex operator that manages three resource types with dependencies:

```
MLPipeline (primary)
    ↓ creates
MLTask (component)
    ↓ creates
MLJob (workload)
```

**Additional Requirements:**
- MLPipeline status should reflect all MLTask statuses
- MLTask status should reflect MLJob completion
- Changes to external ConfigMaps should trigger MLPipeline reconciliation
- Quota changes should trigger re-evaluation of all pipelines

### Task 3.1: Hierarchical Watch Design (5 minutes)

Design watch configurations for each controller:

**MLPipelineController:**
```go
func (r *MLPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Design watches for:
    // 1. Primary MLPipeline resources
    // 2. Owned MLTask resources
    // 3. External ConfigMaps
    // 4. ResourceQuotas
}
```

**MLTaskController:**
```go
func (r *MLTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Design watches for:
    // 1. Primary MLTask resources
    // 2. Owned MLJob resources
}
```

**MLJobController:**
```go
func (r *MLJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Design watches for:
    // 1. Primary MLJob resources
    // 2. Owned Pods/Jobs
}
```

### Task 3.2: Event Propagation Strategy (5 minutes)

Design how status changes propagate up the hierarchy:

1. **MLJob Completion → MLTask Status**:
   - What predicate on MLJob changes should trigger MLTask reconciliation?
   - How should the MLTask controller aggregate MLJob statuses?

2. **MLTask Status → MLPipeline Status**:
   - What MLTask changes should trigger MLPipeline reconciliation?
   - How should the MLPipeline controller aggregate MLTask statuses?

3. **External Resource Changes**:
   - How should ConfigMap changes propagate to MLPipeline reconciliation?
   - What event mapper would connect ResourceQuota changes to all affected MLPipelines?

---

## Exercise 4: Debugging Scenarios (5 minutes)

### Scenario: Watch Troubleshooting

You're debugging watch-related issues in your operator. For each problem, identify the likely cause and solution:

### Problem 1: Missing Reconciliation
**Symptom**: MLPlatform resources aren't reconciling when their ConfigMaps change
**Current Setup**:
```go
Watches(&corev1.ConfigMap{},
    reconciler.WithEventHandler(handler.EnqueueRequestForOwner{}))
```
**Question**: What's wrong and how would you fix it?

### Problem 2: Excessive Reconciliation
**Symptom**: Reconciler runs constantly, even when nothing meaningful changes
**Current Setup**:
```go
Owns(&appsv1.Deployment{})  // No predicates
```
**Question**: What predicate would reduce unnecessary reconciliation?

### Problem 3: Status Update Loops
**Symptom**: Controller logs show continuous reconciliation with no external changes
**Possible Causes**: List 3 potential causes and solutions

### Problem 4: Cross-Resource Issues
**Symptom**: Changes to Node resources aren't triggering MLPlatform reconciliation
**Current Setup**:
```go
Watches(&corev1.Node{},
    reconciler.WithEventHandler(handler.EnqueueRequestForObject{}))
```
**Question**: What's wrong with this event handler choice?

---

## Reflection Questions (5 minutes)

After completing the exercises, reflect on these questions:

1. **Design Trade-offs**: What are the main trade-offs between watching more vs fewer resources?

2. **Predicate Complexity**: When does predicate logic become too complex? How do you balance filtering efficiency with maintainability?

3. **Performance Patterns**: What patterns did you identify for optimizing watch performance in high-traffic environments?

4. **Debugging Strategy**: How would you systematically debug watch-related issues in a production environment?

5. **Best Practices**: Based on the exercises, what are your top 3 best practices for designing watch configurations?

---

## Answer Key and Discussion Points

### Exercise 1 Solutions

**Basic Watch Setup:**
```go
func (r *MLPlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&mlv1.MLPlatform{}).                    // Primary resource
        Owns(&appsv1.Deployment{}).                 // Direct ownership
        Owns(&corev1.Service{}).                    // Direct ownership
        Owns(&corev1.ConfigMap{}).                  // Direct ownership
        Owns(&corev1.Secret{}).                     // Direct ownership
        Watches(&corev1.Node{},                     // External resource
            reconciler.WithEventHandler(/* custom handler */)).
        Watches(&corev1.ResourceQuota{},            // External resource
            reconciler.WithEventHandler(/* custom handler */)).
        Watches(&corev1.PersistentVolumeClaim{},    // May not be owned
            reconciler.WithEventHandler(/* custom handler */)).
        Complete(r)
}
```

**Key Insights:**
- Use `.Owns()` for resources created and managed by the controller
- Use `.Watches()` for external resources that influence behavior
- Consider whether resources are truly "owned" or just "influenced by"

### Exercise 2 Optimization Results

Proper predicate design can typically reduce reconciliation by 70-90% in high-traffic environments by filtering out irrelevant events.

### Exercise 3 Architecture Insights

Hierarchical controllers require careful event mapper design to propagate changes efficiently without creating reconciliation storms.

These exercises prepare you to design efficient, maintainable watch configurations for real-world operator scenarios.