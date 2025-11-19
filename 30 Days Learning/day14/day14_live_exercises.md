# Day 14: Week 2 Review and Controller Exercise - Live Exercises

## Exercise Time: 40 minutes

These hands-on exercises walk you through building the TaskRunner controller from scratch, applying all the patterns and concepts from Week 2. You'll create a working controller that demonstrates mastery of Kubebuilder, reconciliation, and ODH patterns.

---

## Exercise 1: Project Setup and API Design (12 minutes)

### Goal
Set up a Kubebuilder project and define the TaskRunner API using patterns learned from ODH.

### Step 1.1: Initialize Kubebuilder Project (4 minutes)

1. **Create a new directory for your controller**:
   ```bash
   mkdir taskrunner-operator
   cd taskrunner-operator
   ```

2. **Initialize the Kubebuilder project**:
   ```bash
   # Initialize with Go modules
   kubebuilder init --domain example.com --repo github.com/example/taskrunner-operator

   # Create the API
   kubebuilder create api --group batch --version v1 --kind TaskRunner --resource --controller
   ```

3. **Verify the project structure**:
   ```bash
   # Check generated files
   ls -la api/v1/
   ls -la internal/controller/

   # View the scaffolded files
   cat api/v1/taskrunner_types.go | head -20
   ```

### Step 1.2: Define the TaskRunner API (8 minutes)

1. **Edit the TaskRunner types** (`api/v1/taskrunner_types.go`):
   ```go
   // Replace the spec and status structs with:

   // TaskRunnerSpec defines the desired state of TaskRunner
   type TaskRunnerSpec struct {
       // Command to execute
       // +kubebuilder:validation:Required
       // +kubebuilder:validation:MinItems=1
       Command []string `json:"command"`

       // Container image to run
       // +kubebuilder:validation:Required
       Image string `json:"image"`

       // Number of parallel executions
       // +kubebuilder:default=1
       // +kubebuilder:validation:Minimum=1
       // +kubebuilder:validation:Maximum=100
       Parallelism int32 `json:"parallelism,omitempty"`

       // Completion deadline in seconds
       // +kubebuilder:validation:Minimum=1
       DeadlineSeconds *int64 `json:"deadlineSeconds,omitempty"`

       // Management state following ODH pattern
       // +kubebuilder:default=Managed
       // +kubebuilder:validation:Enum=Managed;Unmanaged;Removed
       ManagementState string `json:"managementState,omitempty"`
   }

   // TaskRunnerStatus defines the observed state of TaskRunner
   type TaskRunnerStatus struct {
       // Current phase of the task
       // +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
       Phase string `json:"phase,omitempty"`

       // Number of active jobs
       Active int32 `json:"active,omitempty"`

       // Number of successful completions
       Succeeded int32 `json:"succeeded,omitempty"`

       // Number of failed attempts
       Failed int32 `json:"failed,omitempty"`

       // Conditions represent the latest available observations
       Conditions []metav1.Condition `json:"conditions,omitempty"`
   }
   ```

2. **Add phase constants**:
   ```go
   // Add these constants after the structs
   const (
       TaskPhasePending   = "Pending"
       TaskPhaseRunning   = "Running"
       TaskPhaseSucceeded = "Succeeded"
       TaskPhaseFailed    = "Failed"

       // Management states
       ManagementStateManaged   = "Managed"
       ManagementStateUnmanaged = "Unmanaged"
       ManagementStateRemoved   = "Removed"

       // Finalizer
       TaskRunnerFinalizer = "batch.example.com/finalizer"
   )
   ```

3. **Generate code and manifests**:
   ```bash
   # Generate the updated code
   make generate

   # Generate CRD manifests
   make manifests

   # Check what was generated
   cat config/crd/bases/batch.example.com_taskrunners.yaml | head -30
   ```

---

## Exercise 2: Implement Basic Reconciler Logic (15 minutes)

### Goal
Implement the core reconciliation logic following ODH patterns.

### Step 2.1: Basic Reconciler Structure (7 minutes)

1. **Edit the controller** (`internal/controller/taskrunner_controller.go`):
   ```go
   // Add required imports at the top
   import (
       kbatch "k8s.io/api/batch/v1"
       corev1 "k8s.io/api/core/v1"
       apierrors "k8s.io/apimachinery/pkg/api/errors"
       "k8s.io/apimachinery/pkg/api/meta"
       metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
       "k8s.io/apimachinery/pkg/types"
       "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
       "time"
   )
   ```

2. **Replace the Reconcile method**:
   ```go
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

       log.Info("Reconciling TaskRunner", "phase", taskRunner.Status.Phase, "managementState", taskRunner.Spec.ManagementState)

       // Handle management state
       if taskRunner.Spec.ManagementState == ManagementStateRemoved {
           return r.reconcileRemoved(ctx, taskRunner)
       }

       if taskRunner.Spec.ManagementState == ManagementStateUnmanaged {
           log.Info("TaskRunner is unmanaged, skipping reconciliation")
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

### Step 2.2: Implement Helper Methods (8 minutes)

1. **Add reconciliation helper methods**:
   ```go
   func (r *TaskRunnerReconciler) reconcileNormal(ctx context.Context, taskRunner *batchv1.TaskRunner) (ctrl.Result, error) {
       log := r.Log.WithValues("taskrunner", taskRunner.Name, "namespace", taskRunner.Namespace)

       // Ensure finalizer
       if !controllerutil.ContainsFinalizer(taskRunner, TaskRunnerFinalizer) {
           controllerutil.AddFinalizer(taskRunner, TaskRunnerFinalizer)
           return ctrl.Result{}, r.Update(ctx, taskRunner)
       }

       // Create or get existing job
       job, err := r.createOrUpdateJob(ctx, taskRunner)
       if err != nil {
           log.Error(err, "Failed to create or update job")
           return ctrl.Result{}, err
       }

       // Update status based on job state
       err = r.updateStatus(ctx, taskRunner, job)
       if err != nil {
           log.Error(err, "Failed to update status")
           return ctrl.Result{}, err
       }

       // Determine requeue strategy based on phase
       switch taskRunner.Status.Phase {
       case TaskPhaseSucceeded, TaskPhaseFailed:
           // Requeue less frequently for completed tasks
           return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
       default:
           // Regular monitoring for active tasks
           return ctrl.Result{RequeueAfter: time.Minute}, nil
       }
   }

   func (r *TaskRunnerReconciler) reconcileDelete(ctx context.Context, taskRunner *batchv1.TaskRunner) (ctrl.Result, error) {
       log := r.Log.WithValues("taskrunner", taskRunner.Name, "namespace", taskRunner.Namespace)
       log.Info("Handling TaskRunner deletion")

       // Clean up any resources if needed
       // For this example, Jobs will be cleaned up automatically via owner references

       // Remove finalizer to allow deletion
       controllerutil.RemoveFinalizer(taskRunner, TaskRunnerFinalizer)
       return ctrl.Result{}, r.Update(ctx, taskRunner)
   }

   func (r *TaskRunnerReconciler) reconcileRemoved(ctx context.Context, taskRunner *batchv1.TaskRunner) (ctrl.Result, error) {
       log := r.Log.WithValues("taskrunner", taskRunner.Name, "namespace", taskRunner.Namespace)
       log.Info("TaskRunner marked for removal")

       // Delete associated job
       jobName := taskRunner.Name + "-job"
       job := &kbatch.Job{}
       err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: taskRunner.Namespace}, job)
       if err == nil {
           err = r.Delete(ctx, job)
           if err != nil {
               return ctrl.Result{}, err
           }
       } else if !apierrors.IsNotFound(err) {
           return ctrl.Result{}, err
       }

       return ctrl.Result{RequeueAfter: time.Minute}, nil
   }
   ```

---

## Exercise 3: Implement Job Management (10 minutes)

### Goal
Create the job management logic that creates and monitors Kubernetes Jobs.

### Step 3.1: Job Creation Logic (6 minutes)

1. **Add the job creation method**:
   ```go
   func (r *TaskRunnerReconciler) createOrUpdateJob(ctx context.Context, taskRunner *batchv1.TaskRunner) (*kbatch.Job, error) {
       jobName := taskRunner.Name + "-job"

       // Check if job already exists
       existingJob := &kbatch.Job{}
       err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: taskRunner.Namespace}, existingJob)
       if err == nil {
           // Job exists, return it
           return existingJob, nil
       } else if !apierrors.IsNotFound(err) {
           // Real error occurred
           return nil, err
       }

       // Create new job
       job := &kbatch.Job{
           ObjectMeta: metav1.ObjectMeta{
               Name:      jobName,
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

       // Create the job
       err = r.Create(ctx, job)
       return job, err
   }
   ```

### Step 3.2: Status Update Logic (4 minutes)

1. **Add the status update method**:
   ```go
   func (r *TaskRunnerReconciler) updateStatus(ctx context.Context, taskRunner *batchv1.TaskRunner, job *kbatch.Job) error {
       // Update counters from job status
       taskRunner.Status.Active = job.Status.Active
       taskRunner.Status.Succeeded = job.Status.Succeeded
       taskRunner.Status.Failed = job.Status.Failed

       // Determine current phase
       oldPhase := taskRunner.Status.Phase

       if job.Status.Succeeded > 0 {
           taskRunner.Status.Phase = TaskPhaseSucceeded
       } else if job.Status.Failed > 0 {
           taskRunner.Status.Phase = TaskPhaseFailed
       } else if job.Status.Active > 0 {
           taskRunner.Status.Phase = TaskPhaseRunning
       } else {
           taskRunner.Status.Phase = TaskPhasePending
       }

       // Update conditions when phase changes
       if oldPhase != taskRunner.Status.Phase {
           condition := r.buildConditionForPhase(taskRunner.Status.Phase)
           meta.SetStatusCondition(&taskRunner.Status.Conditions, condition)
       }

       return r.Status().Update(ctx, taskRunner)
   }

   func (r *TaskRunnerReconciler) buildConditionForPhase(phase string) metav1.Condition {
       condition := metav1.Condition{
           Type:               "Ready",
           LastTransitionTime: metav1.NewTime(time.Now()),
       }

       switch phase {
       case TaskPhaseSucceeded:
           condition.Status = metav1.ConditionTrue
           condition.Reason = "TaskCompleted"
           condition.Message = "Task completed successfully"
       case TaskPhaseFailed:
           condition.Status = metav1.ConditionFalse
           condition.Reason = "TaskFailed"
           condition.Message = "Task execution failed"
       case TaskPhaseRunning:
           condition.Status = metav1.ConditionUnknown
           condition.Reason = "TaskRunning"
           condition.Message = "Task is currently running"
       default:
           condition.Status = metav1.ConditionUnknown
           condition.Reason = "TaskPending"
           condition.Message = "Task is pending execution"
       }

       return condition
   }
   ```

---

## Exercise 4: Configure Watch Patterns (3 minutes)

### Goal
Set up efficient watch patterns using predicates from Day 13.

### Step 4.1: Update SetupWithManager

1. **Replace the SetupWithManager method**:
   ```go
   func (r *TaskRunnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
       return ctrl.NewControllerManagedBy(mgr).
           For(&batchv1.TaskRunner{}).
           Owns(&kbatch.Job{}, reconciler.WithPredicates(r.jobPredicate())).
           Complete(r)
   }

   // Add import for reconciler and predicate
   import (
       reconciler "sigs.k8s.io/controller-runtime/pkg/reconcile"
       "sigs.k8s.io/controller-runtime/pkg/predicate"
       "sigs.k8s.io/controller-runtime/pkg/event"
   )

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

## Exercise 5: Testing Your Controller (Optional)

### Goal
Test your controller implementation with real resources.

### Step 5.1: Build and Test

1. **Build the controller**:
   ```bash
   # Generate and build
   make generate
   make manifests
   make build
   ```

2. **Install CRDs** (if you have a cluster):
   ```bash
   make install
   ```

3. **Create a test TaskRunner**:
   ```yaml
   # test-taskrunner.yaml
   apiVersion: batch.example.com/v1
   kind: TaskRunner
   metadata:
     name: test-task
     namespace: default
   spec:
     command: ["echo", "Hello from TaskRunner!"]
     image: "busybox:latest"
     parallelism: 1
     deadlineSeconds: 300
     managementState: "Managed"
   ```

4. **Apply and test** (if cluster available):
   ```bash
   kubectl apply -f test-taskrunner.yaml
   kubectl get taskrunner test-task -o yaml
   kubectl get job test-task-job -o yaml
   ```

---

## Review and Reflection Questions

After completing the exercises:

1. **Week 2 Integration**: How many concepts from Days 8-13 did you use in this controller?

2. **ODH Patterns**: Which ODH patterns did you successfully implement?

3. **Best Practices**: What best practices from Week 2 are evident in your code?

4. **Improvements**: What would you add to make this controller production-ready?

5. **Understanding**: What was the most challenging part of the implementation?

## Key Accomplishments

By completing this exercise, you've demonstrated:
- ✅ Kubebuilder project setup and API design
- ✅ Proper reconciler implementation with error handling
- ✅ Status management and condition tracking
- ✅ Job creation and lifecycle management
- ✅ Watch optimization with predicates
- ✅ ODH pattern adoption (management states, finalizers)

This hands-on controller serves as proof of your Week 2 mastery and prepares you for the advanced topics in Week 3!