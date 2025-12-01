# Day 14: Week 2 Review and Controller Exercise - Offline Exercises

## Exercise Time: 40-50 minutes

These exercises help you consolidate Week 2 knowledge through design scenarios and theoretical applications. Perfect for practicing controller design concepts without requiring a live Kubernetes environment.

---

## Exercise 1: Controller Design Comparison Analysis (15 minutes)

### Scenario: Multi-Controller Architecture Review

You're tasked with reviewing three different controller architectures. Analyze each approach and identify which Week 2 concepts they demonstrate.

### Task 1.1: Architecture Analysis (8 minutes)

For each architecture, identify the Week 2 patterns used:

#### Architecture A: Simple Resource Controller
```go
// Resource manages StatefulSets directly
func (r *SimpleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Fetch resource
    resource := &v1.SimpleResource{}
    if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Create StatefulSet
    statefulset := buildStatefulSet(resource)
    if err := r.Create(ctx, statefulset); err != nil && !errors.IsAlreadyExists(err) {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

func (r *SimpleReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&v1.SimpleResource{}).
        Owns(&appsv1.StatefulSet{}).
        Complete(r)
}
```

#### Architecture B: ODH-Style Component Controller
```go
// Resource manages components with management states
func (r *ComponentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    instance := &v1.ComponentResource{}
    err := r.Get(ctx, req.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // Handle management state
    switch instance.Spec.ManagementState {
    case "Removed":
        return r.reconcileRemoved(ctx, instance)
    case "Unmanaged":
        return ctrl.Result{}, nil
    default:
        return r.reconcileManaged(ctx, instance)
    }
}

func (r *ComponentReconciler) reconcileManaged(ctx context.Context, instance *v1.ComponentResource) (ctrl.Result, error) {
    // Ensure finalizer
    if !controllerutil.ContainsFinalizer(instance, componentFinalizer) {
        controllerutil.AddFinalizer(instance, componentFinalizer)
        return ctrl.Result{}, r.Update(ctx, instance)
    }

    // Deploy components
    for _, component := range instance.Spec.Components {
        if err := r.deployComponent(ctx, instance, component); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Update status conditions
    condition := metav1.Condition{
        Type:   "Ready",
        Status: metav1.ConditionTrue,
        Reason: "ComponentsDeployed",
    }
    meta.SetStatusCondition(&instance.Status.Conditions, condition)

    return ctrl.Result{RequeueAfter: time.Minute * 5}, r.Status().Update(ctx, instance)
}
```

#### Architecture C: Hierarchical Controller System
```go
// Parent controller watches children with predicates
func (r *ParentReconciler) SetupWithManager(mgr ctrl.Manager) error {
    childPredicate := predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            oldChild := e.ObjectOld.(*v1.ChildResource)
            newChild := e.ObjectNew.(*v1.ChildResource)
            return !reflect.DeepEqual(oldChild.Status, newChild.Status)
        },
    }

    return ctrl.NewControllerManagedBy(mgr).
        For(&v1.ParentResource{}).
        Owns(&v1.ChildResource{}, reconciler.WithPredicates(childPredicate)).
        Watches(&corev1.ConfigMap{},
            reconciler.WithEventMapper(r.mapConfigMapToParent),
            reconciler.WithPredicates(r.configMapPredicate())).
        Complete(r)
}
```

**Analysis Questions:**
- Which architecture follows ODH patterns most closely?
- What Week 2 concepts are missing from Architecture A?
- How does Architecture C demonstrate Day 13 watch patterns?

### Task 1.2: Improvement Recommendations (7 minutes)

For each architecture, suggest improvements using Week 2 best practices:

**Architecture A Improvements:**
- Missing patterns: ________________
- Recommended additions: ___________

**Architecture B Improvements:**
- Missing patterns: ________________
- Recommended additions: ___________

**Architecture C Improvements:**
- Missing patterns: ________________
- Recommended additions: ___________

---

## Exercise 2: Reconciler Logic Design Challenge (20 minutes)

### Scenario: Multi-Tenant Database Controller

Design a controller for a `DatabaseCluster` resource that manages PostgreSQL clusters in a multi-tenant environment.

**Requirements:**
- Support multiple database instances per cluster
- Handle user management and database creation
- Manage backups and monitoring
- Support different management states (ODH pattern)
- Handle graceful scaling and updates

### Task 2.1: API Design (8 minutes)

Design the `DatabaseClusterSpec` and `DatabaseClusterStatus`:

```go
type DatabaseClusterSpec struct {
    // Your design here
}

type DatabaseClusterStatus struct {
    // Your design here
}

// Add any constants or enums needed
```

**Design Considerations:**
- What fields are required vs optional?
- How would you implement management states?
- What status information is most important?
- How would you handle database credentials?

### Task 2.2: Reconciler Flow Design (12 minutes)

Design the reconciliation logic for different scenarios:

#### Scenario A: New Cluster Creation
```go
func (r *DatabaseClusterReconciler) reconcileNewCluster(ctx context.Context, cluster *v1.DatabaseCluster) (ctrl.Result, error) {
    // Step 1: ________________
    // Step 2: ________________
    // Step 3: ________________
    // Return: ________________
}
```

#### Scenario B: Cluster Scaling
```go
func (r *DatabaseClusterReconciler) reconcileScaling(ctx context.Context, cluster *v1.DatabaseCluster) (ctrl.Result, error) {
    // Step 1: ________________
    // Step 2: ________________
    // Step 3: ________________
    // Return: ________________
}
```

#### Scenario C: Backup Management
```go
func (r *DatabaseClusterReconciler) reconcileBackups(ctx context.Context, cluster *v1.DatabaseCluster) (ctrl.Result, error) {
    // Step 1: ________________
    // Step 2: ________________
    // Step 3: ________________
    // Return: ________________
}
```

**Questions to Address:**
- How would you handle database initialization?
- What error recovery strategies would you implement?
- How would you manage database user credentials securely?
- What status conditions would you track?

---

## Exercise 3: Watch Pattern Optimization (10 minutes)

### Scenario: High-Traffic Controller Environment

Your `DatabaseCluster` controller is deployed in a high-traffic environment with:
- 100+ database clusters
- Frequent ConfigMap/Secret updates
- Regular Pod restarts and updates
- Persistent volume changes

### Task 3.1: Watch Configuration Design (5 minutes)

Design the `SetupWithManager` method with optimized watch patterns:

```go
func (r *DatabaseClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        // Primary resource
        For(&v1.DatabaseCluster{}).

        // Owned resources - design with predicates
        // Owns(?, reconciler.WithPredicates(?))

        // External resources - design with event mappers
        // Watches(?, reconciler.WithEventMapper(?))

        Complete(r)
}
```

### Task 3.2: Predicate Design (5 minutes)

Design predicates for these scenarios:

#### StatefulSet Predicate
```go
func (r *DatabaseClusterReconciler) statefulSetPredicate() predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            // What changes should trigger reconciliation?
            // oldSts := e.ObjectOld.(*appsv1.StatefulSet)
            // newSts := e.ObjectNew.(*appsv1.StatefulSet)

            return // Your logic here
        },
    }
}
```

#### Secret Predicate
```go
func (r *DatabaseClusterReconciler) secretPredicate() predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            // Only database-related secrets should trigger reconciliation

            return // Your logic here
        },
    }
}
```

#### PVC Predicate
```go
func (r *DatabaseClusterReconciler) pvcPredicate() predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            // What PVC changes matter for database clusters?

            return // Your logic here
        },
    }
}
```

---

## Exercise 4: Week 2 Integration Assessment (15 minutes)

### Task 4.1: Concept Integration Matrix (8 minutes)

Fill out how each Week 2 concept applies to the DatabaseCluster controller:

| Week 2 Concept | Application in DatabaseCluster | Example Implementation |
|----------------|--------------------------------|------------------------|
| **Kubebuilder Markers** | | |
| **Controller-Runtime Setup** | | |
| **ODH Management States** | | |
| **Reconciler Patterns** | | |
| **Status Conditions** | | |
| **Watch Predicates** | | |
| **Event Filtering** | | |
| **Finalizer Management** | | |
| **Error Handling** | | |
| **Owner References** | | |

### Task 4.2: Best Practices Checklist (7 minutes)

For your DatabaseCluster design, check which best practices you've incorporated:

#### Kubebuilder Best Practices
- [ ] Appropriate validation markers on spec fields
- [ ] Clear default values where needed
- [ ] Proper enum validation for state fields
- [ ] Structured status with conditions

#### Controller Best Practices
- [ ] Defensive resource fetching (handle NotFound)
- [ ] Proper finalizer management
- [ ] Separation of reconciliation concerns
- [ ] Appropriate requeue strategies
- [ ] Structured error handling and logging

#### ODH Pattern Adoption
- [ ] Management state handling (Managed/Unmanaged/Removed)
- [ ] Status condition management with proper transitions
- [ ] Component lifecycle patterns
- [ ] Resource cleanup on removal

#### Watch Optimization
- [ ] Predicates to filter unnecessary events
- [ ] Event mappers for cross-resource relationships
- [ ] Appropriate watch scope (Owns vs Watches)
- [ ] Performance considerations for high-traffic environments

---

## Exercise 5: Design Review and Reflection (5 minutes)

### Self-Assessment Questions

Answer these questions to assess your Week 2 mastery:

1. **Kubebuilder Workflow**: Can you explain the complete flow from `kubebuilder init` to a working controller?

2. **ODH Patterns**: Which ODH patterns would you definitely include in any production controller? Why?

3. **Reconciler Design**: What's the most important consideration when designing reconciler logic?

4. **Watch Optimization**: When would you choose `.Owns()` vs `.Watches()`? Give specific examples.

5. **Status Management**: How do you decide which status conditions to include in a custom resource?

6. **Error Handling**: What's your strategy for handling transient vs permanent errors in reconcilers?

7. **Performance**: What are the top 3 performance considerations for controllers in production?

### Knowledge Gaps Assessment

Identify areas where you need additional practice:

**Stronger Areas** (concepts you feel confident about):
- ________________________________
- ________________________________
- ________________________________

**Areas for Improvement** (concepts needing more practice):
- ________________________________
- ________________________________
- ________________________________

**Action Plan** for Week 3:
- How will you strengthen weak areas?
- What specific patterns do you want to master?
- Which ODH components will you study more closely?

---

## Answer Key and Discussion Points

### Exercise 1 Solutions

**Architecture Analysis:**
- **Architecture A**: Basic controller pattern, missing finalizers, management states, status conditions
- **Architecture B**: Strong ODH pattern adoption with management states, finalizers, and conditions
- **Architecture C**: Advanced watch patterns with predicates and event mappers

### Exercise 2 Best Practices

**DatabaseCluster API Design should include:**
- Management state enum
- Replica count with validation
- Resource requirements
- Database configuration
- Backup settings
- Network policies

**Reconciler Logic should handle:**
- Initial setup and bootstrapping
- Scaling operations (up and down)
- Configuration updates
- Backup scheduling
- User and permission management
- Health monitoring

### Exercise 3 Watch Optimization

**Effective predicates should:**
- Filter by labels to identify relevant resources
- Only trigger on meaningful status changes
- Consider performance impact of predicate logic
- Use label selectors for efficient filtering

### Week 2 Mastery Indicators

If you completed these exercises successfully, you've demonstrated:
- ✅ Understanding of Kubebuilder workflow and patterns
- ✅ Ability to design complex APIs with proper validation
- ✅ Knowledge of ODH-style controller patterns
- ✅ Skills in reconciler logic design
- ✅ Watch optimization and performance awareness
- ✅ Integration thinking across multiple Week 2 concepts

You're well-prepared for Week 3's advanced topics: CRDs, Webhooks, and OLM!