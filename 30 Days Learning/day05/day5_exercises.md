# Day 5: Client-go Library Basics - Light Exercises

*Exercise Time: 15 minutes*

These exercises focus on **reading and understanding** existing code rather than heavy implementation. Perfect for building comprehension without the complexity of day 4.

## Exercise 1: Code Reading - ODH Client Usage (8 minutes)

### Task: Analyze Client Patterns in ODH

**File to examine**: `/Users/suksubra/Documents/Work/RHOAI/opendatahub-operator/controllers/datasciencecluster_controller.go`

**What to look for** (spend 2 minutes on each):

1. **Client Interface Usage**
   ```bash
   # Search for client operations in the file
   grep -n "r\.Get\|r\.Create\|r\.Update\|r\.List" /path/to/datasciencecluster_controller.go
   ```
   - How does ODH get resources?
   - What pattern is used for error handling?

2. **Resource Type Handling**
   - Find examples of working with different resource types
   - Notice the difference between built-in resources (ConfigMap, Deployment) and custom resources (DataScienceCluster)

3. **Context Usage**
   - How is `context.Context` passed through client operations?
   - Why is context important for client-go operations?

4. **Status Updates**
   - Look for `r.Status().Update()` calls
   - How does ODH separate spec updates from status updates?

**Expected Discovery**: You should see patterns like:
```go
// Get pattern
var dsc dsci.DataScienceCluster
if err := r.Get(ctx, req.NamespacedName, &dsc); err != nil {
    return ctrl.Result{}, client.IgnoreNotFound(err)
}

// Status update pattern
if err := r.Status().Update(ctx, &dsc); err != nil {
    return ctrl.Result{}, err
}
```

## Exercise 2: Pattern Recognition - Informer Setup (4 minutes)

### Task: Understand Informer Configuration

**File to examine**: Look for `SetupWithManager` functions in ODH controllers

**What to analyze**:

1. **Resource Watching**
   ```go
   func (r *SomeReconciler) SetupWithManager(mgr ctrl.Manager) error {
       return ctrl.NewControllerManagedBy(mgr).
           For(&someType{}).     // Primary resource
           Owns(&otherType{}).   // Owned resources
           Complete(r)
   }
   ```

2. **Questions to answer**:
   - What resources does each controller watch?
   - How does `For()` differ from `Owns()`?
   - What happens when you watch multiple resource types?

**Expected Learning**: Each `For()` and `Owns()` call creates an informer that watches for changes to that resource type.

## Exercise 3: Quick Comprehension Check (3 minutes)

### Task: Connect Concepts to Code

**Answer these questions based on your code reading**:

1. **Client Abstraction**
   - Does ODH code directly create clientsets or use controller-runtime's client?
   - What's the advantage of using `r.Get()` vs direct clientset calls?

2. **Error Handling**
   - What does `client.IgnoreNotFound(err)` do and why is it used?
   - When would you NOT want to ignore "not found" errors?

3. **Resource Relationships**
   - How does ODH establish ownership between resources?
   - What happens to owned resources when the owner is deleted?

### Quick Answers (for self-check):

1. **Client**: ODH uses controller-runtime's unified client interface, which abstracts clientset complexity
2. **IgnoreNotFound**: Treats "not found" as success since the resource being deleted/missing might be expected
3. **Ownership**: Uses `controllerutil.SetControllerReference()` to establish parent-child relationships for garbage collection

## Bonus: Quick Reference Creation (Optional)

If you finish early, create a personal quick reference note:

```markdown
## My Client-go Patterns Cheat Sheet

### Common ODH Patterns I Found:
- Get: `r.Get(ctx, namespacedName, &resource)`
- Create: `r.Create(ctx, &resource)`
- Update Status: `r.Status().Update(ctx, &resource)`
- List: `r.List(ctx, &list, client.InNamespace(ns))`

### Key Insights:
- [Your observations from code reading]

### Questions for Day 6:
- [Any questions about controller-runtime that came up]
```

## Exercise Completion

**Time spent**: Should be around 15 minutes total
**Key achievement**: You can now read and understand client-go usage patterns in operator code
**Next step**: Day 6 will show how controller-runtime makes these patterns even easier to use

**No Heavy Lifting**: Notice how these exercises focused on understanding existing patterns rather than writing complex code from scratch. This approach builds comprehension efficiently!