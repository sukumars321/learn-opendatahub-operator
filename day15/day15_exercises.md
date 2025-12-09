# Day 15: Advanced CRD Features - Alternative Exercises

## Overview
These exercises help you practice advanced CRD design patterns without requiring access to a live Kubernetes cluster or the ODH codebase. Focus on understanding concepts, designing schemas, and applying validation patterns.

**Total Time**: 15 minutes
**Prerequisites**: Text editor and understanding of YAML/JSON syntax

---

## Exercise 1: Design a CRD Schema with Advanced Validation (5 minutes)

### Scenario
You're designing a CRD for a **DatabaseCluster** resource that manages distributed database deployments. The resource should support multiple database types with type-specific configuration.

### Requirements
Design a CRD schema that includes:
- Database type selection (PostgreSQL, MySQL, MongoDB)
- Replica count with validation (1-10 range)
- Storage configuration with format validation
- Management state with enum validation
- Optional advanced features with defaults

### Task
Write an OpenAPI v3 schema snippet for the `spec` section:

```yaml
# Your CRD schema here
spec:
  type: object
  properties:
    # Design your schema properties here

    # Include these elements:
    # 1. Database type enum validation
    # 2. Replica count with min/max constraints
    # 3. Storage size with pattern validation (e.g., "10Gi")
    # 4. Management state enum
    # 5. Optional backup configuration with defaults
```

### Example Solution Framework
```yaml
spec:
  type: object
  required: ["databaseType", "managementState"]
  properties:
    databaseType:
      type: string
      enum: ["postgresql", "mysql", "mongodb"]
      description: "Type of database to deploy"

    managementState:
      type: string
      enum: ["Managed", "Unmanaged", "Removed"]
      default: "Managed"

   # add other fields here.....
    replicas:
      type: integer
      minimum: 1
      maximum: 10
      default: 3

   storageSize:
      type: string
      pattern: "^[0-9]+[KMGTPE]i?$"

   version:
      type: string
      pattern: "^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z\-\.]+)?(?:\+[0-9A-Za-z\-\.]+)?$"

   backupEnabled:
      type: boolean
      default: true

     
```

### Validation Points
- [] Enum validation for database types
- [] Numeric constraints for replicas
- [] Pattern validation for storage specifications
- [] Required vs optional field decisions
- [] Meaningful default values

---

## Exercise 2: Kubebuilder Marker Design (4 minutes)

### Scenario
You're writing Go structs for the DatabaseCluster CRD and need to add appropriate kubebuilder markers for validation and code generation.

### Task
Add kubebuilder markers to these Go struct definitions:

```go
// DatabaseClusterSpec defines the desired state of DatabaseCluster
type DatabaseClusterSpec struct {
    // Database type - should be one of: postgresql, mysql, mongodb
    DatabaseType string `json:"databaseType"`

    // Management state - defaults to "Managed"
    ManagementState string `json:"managementState,omitempty"`

    // Number of database replicas - must be between 1 and 10, defaults to 3
    Replicas int32 `json:"replicas,omitempty"`

    // Storage size - must follow Kubernetes quantity format (e.g., "10Gi")
    StorageSize string `json:"storageSize,omitempty"`

    // Database version - must match semantic versioning pattern
    Version string `json:"version,omitempty"`

    // Enable backup - defaults to true
    BackupEnabled bool `json:"backupEnabled,omitempty"`

    // Optional authentication configuration
    Authentication *AuthConfig `json:"authentication,omitempty"`
}
```

### Your Task
Add appropriate `+kubebuilder:` markers above each field. Consider:
- Validation constraints (enums, minimums, patterns)
- Default values
- Documentation
- Optional vs required fields

### Hint Framework
```go
type DatabaseClusterSpec struct {
    // +kubebuilder:validation:Enum=postgresql;mysql;mongodb
    // +kubebuilder:validation:Required
    DatabaseType string `json:"databaseType"`

   // Management state - defaults to "Managed"
   // +kubebuilder:default=Managed
	// +kubebuilder:validation:Enum=Managed;Unmanaged;Removed
    ManagementState string `json:"managementState,omitempty"`

    // Number of database replicas - must be between 1 and 10, defaults to 3
    // +kubebuilder:default=3
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=10
    Replicas int32 `json:"replicas,omitempty"`

    // Storage size - must follow Kubernetes quantity format (e.g., "10Gi")
    // +kubebuilder:validation:Pattern=^[0-9]+[KMGTPE]i?$
    StorageSize string `json:"storageSize,omitempty"`

    // Database version - must match semantic versioning pattern
    // +kubebuilder:validation:Pattern=^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z\-\.]+)?(?:\+[0-9A-Za-z\-\.]+)?$
    Version string `json:"version,omitempty"`

    // Enable backup - defaults to true
    // +kubebuilder:default=true
    BackupEnabled bool `json:"backupEnabled,omitempty"`

    // Optional authentication configuration
    // +kubebuilder:validation:Optional
    Authentication *AuthConfig `json:"authentication,omitempty"`
}
```

### Check Your Understanding
- [ ] Used enum validation for restricted values
- [ ] Set appropriate default values
- [ ] Added numeric constraints where needed
- [ ] Used pattern validation for formatted strings
- [ ] Marked required vs optional fields correctly

---

## Exercise 3: Status Subresource and Conditions Design (3 minutes)

### Scenario
Design the status structure and conditions for your DatabaseCluster CRD. The status should track overall health, individual component readiness, and operational conditions.

### Task 1: Design Status Structure (1.5 minutes)
```go
// DatabaseClusterStatus defines the observed state of DatabaseCluster
type DatabaseClusterStatus struct {
    // Add your status fields here

    // Standard condition pattern
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Operational status
    Phase string `json:"phase,omitempty"`

    ReplicaCount string `json:"replicaCount"`
    
    DatabaseVersion string `json:"databaseVersion"`

    // Should include:
    // - Standard Kubernetes conditions
    // - Current replica count
    // - Database version in use
    // - Connection information
    // - Component-specific status
}
```

### Task 2: Define Condition Types (1.5 minutes)
List the condition types your DatabaseCluster should track:

```go
const (
    // Define condition type constants
    // Examples:
   //  DatabaseReady = "DatabaseReady"
   //  ReplicationHealthy = "ReplicationHealthy"
   //  BackupConfigured = "BackupConfigured"
    DatabaseProgressing = "DatabaseProgressing"
    DatabaseStartupFailed = "DatabaseStartupFailed"
    BackupInProgress = "BackupInProgress"
    BackupCompleted = "BackupCompleted"
)
```

### Example Status Framework
```go
type DatabaseClusterStatus struct {
    // Standard condition pattern
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Operational status
    Phase string `json:"phase,omitempty"`

    // Add your additional status fields...
}
```

### Design Considerations
- [ ] What information do operators need to see quickly?
- [ ] How can conditions help troubleshoot problems?
- [ ] What kubectl columns would be most useful?
- [ ] How should failure states be represented?

---

## Exercise 4: Multi-Version API Design Challenge (3 minutes)

### Scenario
You need to evolve your DatabaseCluster API from v1 to v2. Version 2 adds new features while maintaining backward compatibility.

### Current v1 API Summary
```yaml
spec:
  databaseType: string (enum)
  replicas: integer
  storageSize: string
  managementState: string
```

### v2 Enhancement Requirements
- Add support for custom resource limits (CPU, memory)
- Add support for multiple storage classes
- Add monitoring configuration
- Add network policy settings
- Maintain compatibility with v1

### Task 1: Design v2 Schema Additions (2 minutes)
```yaml
# v2 additions to spec (design only - no need to write full schema)

# New fields for v2:
# 1. resources (CPU, memory limits)
# 2. storage classes and multiple volumes
# 3. monitoring configuration
# 4. network policies

# Your design here...
```

### Task 2: Conversion Strategy (1 minute)
Choose and justify your conversion approach:

**Option A: Schema-Compatible (conversion: strategy: None)**
- Pros:
- Cons:
- When to use:

**Option B: Webhook Conversion (conversion: strategy: Webhook)**
- Pros:
- Cons:
- When to use:

**Your Choice**: _________ **because**: _________________

### Migration Planning Questions
- [ ] Can v1 resources work with v2 controllers?
- [ ] How do new v2 fields get default values?
- [ ] What happens to v1 resources when storage version changes?
- [ ] How do you handle deprecated features?

---

## Conceptual Review Questions

Test your understanding with these questions:

### Schema Validation
1. **When should you use `required` vs `default` for a field?**
   - Required when: _______________
   - Default when: _______________

2. **What's the difference between enum validation and pattern validation?**
   - Enum: _______________
   - Pattern: _______________

3. **How do kubebuilder markers relate to OpenAPI schema?**
   - Markers generate: _______________
   - Schema provides: _______________

### Subresources
1. **Why separate spec and status with subresources?**
   - RBAC: _______________
   - Workflow: _______________
   - Concurrency: _______________

2. **What makes a good kubectl additional printer column?**
   - Content: _______________
   - JSONPath: _______________
   - User value: _______________

### Version Management
1. **What's the difference between served and storage versions?**
   - Served: _______________
   - Storage: _______________

2. **When would you use webhook conversion vs. schema-compatible versioning?**
   - Webhook when: _______________
   - Schema-compatible when: _______________

---

## Design Pattern Analysis

### Real-World Scenarios

**Scenario A**: Your CRD manages ephemeral development environments
- What validation rules matter most?
- How should defaults be optimized?
- What status information helps developers?

**Scenario B**: Your CRD manages production database clusters
- How does validation differ from Scenario A?
- What additional status tracking is needed?
- How should version migration be handled?

**Scenario C**: Your CRD is part of a platform used by multiple teams
- How do defaults balance simplicity vs. flexibility?
- What validation prevents common mistakes?
- How should status support troubleshooting?

---

## Key Takeaways Documentation

As you complete these exercises, document your insights:

### Validation Strategy
- When to use different validation types
- How to balance strictness with usability
- Common validation patterns and their purposes

### Default Value Philosophy
- What should have defaults vs. be required
- How defaults affect user experience
- Strategies for environment-appropriate defaults

### Status Design Principles
- Essential vs. nice-to-have status information
- Condition types that aid troubleshooting
- kubectl integration considerations

### Version Evolution Planning
- Backward compatibility strategies
- Schema design for extensibility
- Migration path considerations

These exercises reinforce advanced CRD concepts through practical design challenges, helping you think like a CRD architect while understanding the trade-offs involved in production API design.