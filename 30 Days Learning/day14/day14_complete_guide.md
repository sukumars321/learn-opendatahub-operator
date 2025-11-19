# Day 14: Week 2 Review and Controller Exercise - Complete Study Guide

## Study Time: 60 minutes (20 min review + 40 min practice)

## Learning Objectives
- Consolidate understanding of Kubebuilder workflow and controller patterns from Week 2
- Build a complete controller from scratch applying ODH patterns
- Practice reconciliation logic, status management, and watch configurations
- Validate understanding through hands-on implementation

---

## Part 1: Week 2 Review and Consolidation (20 minutes)

### Week 2 Journey Recap

Let's quickly review the key concepts and patterns you've learned:

#### Day 8-9: Kubebuilder Foundation
```bash
# Core Kubebuilder workflow you learned
kubebuilder init --domain example.com --repo github.com/example/operator
kubebuilder create api --group apps --version v1 --kind MyApp
make generate  # Generate code from markers
make manifests # Generate YAML manifests
```

**Key Takeaways:**
- Kubebuilder automates controller scaffolding
- Markers drive code generation
- Clear separation between API definitions and controller logic

#### Day 10-11: ODH Architecture Patterns
**Component Management Pattern** (from ODH):
```go
// Pattern observed in ODH components
type ComponentInterface interface {
    ReconcileComponent(ctx context.Context, owner metav1.Object) error
    GetComponentName() string
    GetManagementState() operatorv1.ManagementState
}
```

**Status Management Pattern**:
```go
// ODH status condition pattern
condition := &metav1.Condition{
    Type:    "Ready",
    Status:  metav1.ConditionTrue,
    Reason:  "ComponentDeployed",
    Message: "Component successfully deployed",
}
meta.SetStatusCondition(&instance.Status.Conditions, *condition)
```

#### Day 12: Reconciler Implementation
**Core Reconciler Pattern** (from ODH):
```go
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch resource
    instance := &v1.MyResource{}
    err := r.Get(ctx, req.NamespacedName, instance)
    if err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil // Resource deleted
        }
        return ctrl.Result{}, err
    }

    // 2. Handle deletion
    if instance.DeletionTimestamp != nil {
        return r.reconcileDelete(ctx, instance)
    }

    // 3. Normal reconciliation
    return r.reconcileNormal(ctx, instance)
}
```

#### Day 13: Watch Patterns
**Sophisticated Watch Configuration** (from ODH):
```go
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&v1.MyResource{}).
        Owns(&appsv1.Deployment{}, reconciler.WithPredicates(deploymentPredicate)).
        Watches(&corev1.ConfigMap{}, reconciler.WithEventMapper(configMapMapper)).
        Complete(r)
}
```

### Quick Self-Assessment Questions

Before proceeding, answer these to gauge your understanding:

1. **Kubebuilder**: What command generates controller scaffolding for a new API?
2. **Controllers**: What are the three main phases of a reconciliation loop?
3. **ODH Patterns**: How does ODH handle component management states?
4. **Status**: What's the difference between spec and status in custom resources?
5. **Watches**: When would you use `.Owns()` vs `.Watches()`?

---

## Part 2: Hands-On Controller Development (40 minutes)

### Project: Build a "TaskRunner" Controller

You'll build a controller for a `TaskRunner` custom resource that manages batch jobs with monitoring and cleanup capabilities.

#### Step 1: Design the API (10 minutes)

**TaskRunner Specification:**
```go
// TaskRunnerSpec defines the desired state
type TaskRunnerSpec struct {
    // Command to execute
    Command []string `json:"command"`

    // Container image
    Image string `json:"image"`

    // Number of parallel executions
    Parallelism int32 `json:"parallelism,omitempty"`

    // Completion deadline in seconds
    DeadlineSeconds *int64 `json:"deadlineSeconds,omitempty"`

    // Management state (following ODH pattern)
    ManagementState operatorv1.ManagementState `json:"managementState,omitempty"`
}

// TaskRunnerStatus defines the observed state
type TaskRunnerStatus struct {
    // Current phase of the task
    Phase TaskPhase `json:"phase,omitempty"`

    // Number of active jobs
    Active int32 `json:"active,omitempty"`

    // Number of successful completions
    Succeeded int32 `json:"succeeded,omitempty"`

    // Number of failed attempts
    Failed int32 `json:"failed,omitempty"`

    // Conditions represent the latest available observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type TaskPhase string

const (
    TaskPhasePending   TaskPhase = "Pending"
    TaskPhaseRunning   TaskPhase = "Running"
    TaskPhaseSucceeded TaskPhase = "Succeeded"
    TaskPhaseFailed    TaskPhase = "Failed"
)
```

**Design Questions to Consider:**
- What validation should be added to the spec?
- What conditions should be tracked in status?
- How should the controller handle different management states?

#### Step 2: Implement Core Reconciler Logic (15 minutes)

**Reconciler Structure** (following ODH patterns):
```go
type TaskRunnerReconciler struct {
    client.Client
    Log    logr.Logger
    Scheme *runtime.Scheme
}

func (r *TaskRunnerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("taskrunner", req.NamespacedName)

    // Fetch the TaskRunner instance
    taskRunner := &batchv1.TaskRunner{}
    err := r.Get(ctx, req.NamespacedName, taskRunner)
    if err != nil {
        if apierrors.IsNotFound(err) {
            log.Info("TaskRunner resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        log.Error(err, "Failed to get TaskRunner")
        return ctrl.Result{}, err
    }

    // Handle management state (ODH pattern)
    if taskRunner.Spec.ManagementState == operatorv1.Removed {
        return r.reconcileRemoved(ctx, taskRunner)
    }

    if taskRunner.Spec.ManagementState == operatorv1.Unmanaged {
        return ctrl.Result{}, nil
    }

    // Handle deletion with finalizers
    if taskRunner.ObjectMeta.DeletionTimestamp != nil {
        return r.reconcileDelete(ctx, taskRunner)
    }

    // Normal reconciliation
    return r.reconcileNormal(ctx, taskRunner)
}
```

**Key Implementation Functions:**
```go
func (r *TaskRunnerReconciler) reconcileNormal(ctx context.Context, taskRunner *batchv1.TaskRunner) (ctrl.Result, error) {
    // 1. Ensure finalizer
    if !controllerutil.ContainsFinalizer(taskRunner, taskRunnerFinalizer) {
        controllerutil.AddFinalizer(taskRunner, taskRunnerFinalizer)
        return ctrl.Result{}, r.Update(ctx, taskRunner)
    }

    // 2. Create or update Job
    job, err := r.createOrUpdateJob(ctx, taskRunner)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 3. Update status based on job state
    err = r.updateStatus(ctx, taskRunner, job)
    if err != nil {
        return ctrl.Result{}, err
    }

    // 4. Handle completion or failure
    if taskRunner.Status.Phase == TaskPhaseSucceeded || taskRunner.Status.Phase == TaskPhaseFailed {
        // Cleanup if needed, requeue for monitoring
        return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
    }

    // 5. Regular monitoring
    return ctrl.Result{RequeueAfter: time.Minute}, nil
}
```

#### Step 3: Implement Job Management (10 minutes)

**Job Creation Logic:**
```go
func (r *TaskRunnerReconciler) createOrUpdateJob(ctx context.Context, taskRunner *batchv1.TaskRunner) (*kbatch.Job, error) {
    // Define desired Job
    job := &kbatch.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      taskRunner.Name + "-job",
            Namespace: taskRunner.Namespace,
            Labels: map[string]string{
                "app.kubernetes.io/name":       "taskrunner",
                "app.kubernetes.io/instance":   taskRunner.Name,
                "app.kubernetes.io/created-by": "taskrunner-controller",
            },
        },
        Spec: kbatch.JobSpec{
            Parallelism:           &taskRunner.Spec.Parallelism,
            ActiveDeadlineSeconds: taskRunner.Spec.DeadlineSeconds,
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name:    "task",
                            Image:   taskRunner.Spec.Image,
                            Command: taskRunner.Spec.Command,
                        },
                    },
                },
            },
        },
    }

    // Set TaskRunner as owner
    if err := controllerutil.SetControllerReference(taskRunner, job, r.Scheme); err != nil {
        return nil, err
    }

    // Create or update
    found := &kbatch.Job{}
    err := r.Get(ctx, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
    if err != nil && apierrors.IsNotFound(err) {
        err = r.Create(ctx, job)
        return job, err
    } else if err != nil {
        return nil, err
    }

    return found, nil
}
```

**Status Update Logic:**
```go
func (r *TaskRunnerReconciler) updateStatus(ctx context.Context, taskRunner *batchv1.TaskRunner, job *kbatch.Job) error {
    // Update counters
    taskRunner.Status.Active = job.Status.Active
    taskRunner.Status.Succeeded = job.Status.Succeeded
    taskRunner.Status.Failed = job.Status.Failed

    // Determine phase
    if job.Status.Succeeded > 0 {
        taskRunner.Status.Phase = TaskPhaseSucceeded
    } else if job.Status.Failed > 0 {
        taskRunner.Status.Phase = TaskPhaseFailed
    } else if job.Status.Active > 0 {
        taskRunner.Status.Phase = TaskPhaseRunning
    } else {
        taskRunner.Status.Phase = TaskPhasePending
    }

    // Update conditions (following ODH pattern)
    condition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionUnknown,
        LastTransitionTime: metav1.NewTime(time.Now()),
        Reason:             "Reconciling",
        Message:            "TaskRunner is being processed",
    }

    if taskRunner.Status.Phase == TaskPhaseSucceeded {
        condition.Status = metav1.ConditionTrue
        condition.Reason = "TaskCompleted"
        condition.Message = "Task completed successfully"
    } else if taskRunner.Status.Phase == TaskPhaseFailed {
        condition.Status = metav1.ConditionFalse
        condition.Reason = "TaskFailed"
        condition.Message = "Task execution failed"
    }

    meta.SetStatusCondition(&taskRunner.Status.Conditions, condition)

    return r.Status().Update(ctx, taskRunner)
}
```

#### Step 4: Configure Watch Patterns (5 minutes)

**Controller Setup** (applying Day 13 patterns):
```go
func (r *TaskRunnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&batchv1.TaskRunner{}).
        Owns(&kbatch.Job{}, reconciler.WithPredicates(r.jobPredicate())).
        Complete(r)
}

func (r *TaskRunnerReconciler) jobPredicate() predicate.Predicate {
    return predicate.Funcs{
        UpdateFunc: func(e event.UpdateEvent) bool {
            oldJob := e.ObjectOld.(*kbatch.Job)
            newJob := e.ObjectNew.(*kbatch.Job)

            // Only reconcile when job status changes meaningfully
            return oldJob.Status.Active != newJob.Status.Active ||
                   oldJob.Status.Succeeded != newJob.Status.Succeeded ||
                   oldJob.Status.Failed != newJob.Status.Failed
        },
        CreateFunc: func(e event.CreateEvent) bool {
            return true
        },
        DeleteFunc: func(e event.DeleteEvent) bool {
            return true
        },
    }
}
```

---

## Key Learning Reinforcement

### Patterns Applied From Week 2

1. **Kubebuilder Structure**: Clean API definition with proper Go tags
2. **ODH Management State**: Following the Managed/Unmanaged/Removed pattern
3. **Reconciler Flow**: Fetch → Handle deletion → Normal reconciliation
4. **Status Conditions**: Using metav1.Condition with proper transitions
5. **Watch Optimization**: Predicates to filter unnecessary reconciliation
6. **Error Handling**: Proper error propagation and logging
7. **Owner References**: Establishing resource relationships

### Best Practices Demonstrated

- **Defensive Programming**: Always check for resource existence
- **Finalizer Management**: Proper cleanup on resource deletion
- **Status Updates**: Separate status updates from spec changes
- **Event Filtering**: Use predicates to optimize performance
- **Logging**: Structured logging for observability
- **Resource Ownership**: Clear parent-child relationships

### Common Pitfalls to Avoid

- **Missing Finalizers**: Resources may not clean up properly
- **Status Update Loops**: Updating status should not trigger reconciliation
- **Poor Error Handling**: Always handle and log errors appropriately
- **Inefficient Watches**: Use predicates to filter events
- **Resource Leaks**: Ensure proper cleanup on deletion

This controller exercise demonstrates how all the Week 2 concepts work together to create a functioning Kubernetes operator following ODH patterns and best practices.