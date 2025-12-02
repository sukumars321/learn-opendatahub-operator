/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/example/taskrunner-operator/api/v1"
	"github.com/go-logr/logr"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Add required imports at the top

// TaskRunnerReconciler reconciles a TaskRunner object
type TaskRunnerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

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

// +kubebuilder:rbac:groups=batch.example.com,resources=taskrunners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.example.com,resources=taskrunners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch.example.com,resources=taskrunners/finalizers,verbs=update

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

func (r *TaskRunnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.TaskRunner{}).
		Owns(&kbatch.Job{}, builder.WithPredicates(r.jobPredicate())).
		Complete(r)
}

func (r *TaskRunnerReconciler) jobPredicate() predicate.Funcs {
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
