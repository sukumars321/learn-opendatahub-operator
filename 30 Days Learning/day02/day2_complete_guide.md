# Day 2 Complete Study Guide: Custom Resource Definitions (CRDs) Basics

## 🎯 Day 2 Goal
Understand what CRDs are, how they extend Kubernetes, and explore the sophisticated CRD patterns used in the OpenDataHub Operator.

**Building on Day 1**: You now know how APIs work - today you'll learn how ODH creates entirely new API types!

---

## 📚 Study Topics (40 minutes)

### 1. What are CRDs and Why They Exist (15 minutes)

#### The Problem CRDs Solve:
- **Kubernetes Core**: Limited to built-in resources (pods, services, deployments)
- **Real Applications**: Need domain-specific resources (databases, ML pipelines, data science clusters)
- **Extension Challenge**: How to add new resource types without modifying Kubernetes core?

#### CRDs are the Solution:
- **Custom Resource Definition**: Defines a new API type
- **Custom Resource**: An instance of that new type
- **Controller**: Watches and manages custom resources
- **Result**: Kubernetes becomes programmable and extensible

#### ODH Example:
```yaml
# This wouldn't exist without CRDs:
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: default-dsc
spec:
  components:
    dashboard:
      managementState: Managed
```

#### Key Benefits:
1. **Domain-Specific APIs**: Resources that match your problem domain
2. **Declarative Management**: Same patterns as core Kubernetes
3. **Tool Integration**: `kubectl`, monitoring, RBAC all work automatically
4. **Ecosystem Growth**: Operators, controllers, custom logic

### 2. CRD Structure: Spec, Status, Metadata (15 minutes)

#### The Three-Part Structure:

**Metadata** (Who and Where):
```yaml
metadata:
  name: my-resource           # Unique identifier
  namespace: my-namespace     # Scope (if namespaced)
  labels:                     # Key-value pairs for organization
    app: my-app
  annotations:                # Additional metadata
    description: "My resource"
  ownerReferences:            # Parent-child relationships
  - apiVersion: apps/v1
    kind: Deployment
    name: my-deployment
    uid: abc-123
```

**Spec** (Desired State - What You Want):
```yaml
spec:
  # User-defined desired configuration
  components:
    dashboard:
      managementState: Managed
    workbenches:
      managementState: Managed
  devFlags:
    manifests:
    - uri: "https://my-custom-manifests"
```

**Status** (Current State - What Actually Exists):
```yaml
status:
  # Controller-managed current state
  phase: Ready
  conditions:
  - type: Ready
    status: "True"
    reason: ReconcileCompleted
    lastTransitionTime: "2024-09-29T10:30:00Z"
  - type: Degraded
    status: "False"
    reason: AllComponentsHealthy
```

#### Critical Rules:
- **Users write spec** - declares what they want
- **Controllers write status** - reports what actually exists
- **Never mix them** - spec and status serve different purposes
- **Status is read-only** to users - only controllers update it

### 3. OpenAPI v3 Schema Validation (10 minutes)

#### Why Schema Validation Matters:
- **Type Safety**: Prevents invalid configurations
- **Documentation**: Self-documenting APIs
- **Tool Support**: Auto-completion, validation in IDEs
- **Error Prevention**: Catch mistakes before they cause problems

#### ODH Schema Example:
```yaml
# From DataScienceCluster CRD
schema:
  openAPIV3Schema:
    type: object
    properties:
      spec:
        type: object
        properties:
          components:
            type: object
            properties:
              dashboard:
                type: object
                properties:
                  managementState:
                    type: string
                    enum: ["Managed", "Removed"]
                    default: "Managed"
                required: ["managementState"]
```

#### Schema Features:
- **Type Constraints**: string, integer, boolean, object, array
- **Validation Rules**: required fields, enums, patterns, min/max values
- **Default Values**: Automatically set if not provided
- **Nested Objects**: Complex hierarchical structures
- **Array Validation**: Rules for list items

#### Real-World Impact:
```bash
# This would be rejected by schema validation:
spec:
  components:
    dashboard:
      managementState: "InvalidValue"  # Not in enum ["Managed", "Removed"]
```

---

## 🔬 Hands-on Exercises (20 minutes)

### Exercise 1: Explore ODH CRD Definitions (5 minutes)

```bash
# Get all ODH CRDs
oc get crd | grep opendatahub

# Examine the main DataScienceCluster CRD
oc get crd datascienceclusters.datasciencecluster.opendatahub.io -o yaml > dsc-crd.yaml

# Look at a component CRD
oc get crd dashboards.components.platform.opendatahub.io -o yaml > dashboard-crd.yaml
```

**Analysis Questions:**
1. How many CRDs does ODH define?
2. What's the difference between `datasciencecluster.opendatahub.io` and `components.platform.opendatahub.io` groups?
3. Which resources are cluster-scoped vs namespaced?

### Exercise 2: Examine Live ODH Resources (5 minutes)

```bash
# Check if DataScienceCluster exists
oc get datascienceclusters
oc get dsc -o yaml

# Look at component resources
oc get dashboards -o yaml
oc get workbenches -o yaml

# Examine the structure
oc describe dsc
oc describe dashboard
```

**Learning Points:**
- See real `spec` vs `status` sections
- Notice how `status.conditions` report current state
- Understand owner reference relationships
- Observe how defaults are applied

### Exercise 3: Understand CRD Schema (5 minutes)

```bash
# Use kubectl explain to explore the schema
oc explain datasciencecluster
oc explain datasciencecluster.spec
oc explain datasciencecluster.spec.components
oc explain datasciencecluster.status

# Look at component schemas
oc explain dashboard
oc explain dashboard.spec
oc explain workbenches.spec
```

**Key Observations:**
- How detailed is the schema documentation?
- What fields are required vs optional?
- What validation rules can you identify?
- How are nested objects structured?

### Exercise 4: Create a Simple Custom Resource (5 minutes)

Let's create a simple test resource to understand the process:

```bash
# First, let's see what happens when we try to create a DataScienceCluster
# (Don't actually apply this - just examine it)
cat << 'EOF' > test-dsc.yaml
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:
  name: test-cluster
spec:
  components:
    dashboard:
      managementState: Managed
    workbenches:
      managementState: Removed
EOF

# Validate the resource without creating it
oc apply --dry-run=client -f test-dsc.yaml

# Check what would happen with server-side validation
oc apply --dry-run=server -f test-dsc.yaml
```

**Understanding the Process:**
1. **Client Validation**: Basic YAML structure
2. **Server Validation**: Schema validation against CRD
3. **Admission**: Webhooks can modify or reject
4. **Storage**: If valid, stored in etcd
5. **Controller Reaction**: ODH controller sees the change

---

## 🔍 ODH Code References

### CRD Files to Examine:
```bash
# In the ODH codebase at /Users/suksubra/Documents/Work/RHOAI/opendatahub-operator
/config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
/config/crd/bases/dscinitialization.opendatahub.io_dscinitializations.yaml
/config/crd/bases/components.platform.opendatahub.io_dashboards.yaml
```

### Go Type Definitions:
```bash
# The Go structs that generate these CRDs
/apis/datasciencecluster/v1/datasciencecluster_types.go
/apis/dscinitialization/v1/dscinitialization_types.go
/apis/components/v1alpha1/dashboard_types.go
```

Let's look at how Go types become CRDs:

```go
// From datasciencecluster_types.go
type DataScienceClusterSpec struct {
    Components Components `json:"components,omitempty"`
    DevFlags   DevFlags   `json:"devFlags,omitempty"`
}

type DataScienceClusterStatus struct {
    Phase      string             `json:"phase,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type DataScienceCluster struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   DataScienceClusterSpec   `json:"spec,omitempty"`
    Status DataScienceClusterStatus `json:"status,omitempty"`
}
```

**Key Observations:**
- **Kubebuilder markers** (`+kubebuilder:`) generate CRD YAML
- **JSON tags** define the API field names
- **Go types** provide structure and validation
- **Status subresource** separates spec from status

---

## 🧠 Key Takeaways

### 1. CRDs Enable Extension
- Kubernetes becomes programmable through CRDs
- ODH adds 20+ new resource types without touching Kubernetes core
- Same tools and patterns work for custom resources

### 2. Three-Part Structure is Universal
- **Metadata**: Identity and relationships
- **Spec**: User's desired state
- **Status**: Controller's current state reporting

### 3. Schema Validation Provides Safety
- OpenAPI v3 schemas prevent configuration errors
- Rich validation rules (types, enums, patterns, required fields)
- Self-documenting APIs through schema descriptions

### 4. Controllers Make It Work
- CRDs define the API, controllers implement the behavior
- Controllers watch custom resources and reconcile state
- Status conditions provide rich feedback to users

### 5. ODH's Sophisticated Patterns
- **Hierarchical CRDs**: DataScienceCluster manages components
- **Multiple API groups**: Logical organization of functionality
- **Version evolution**: v1alpha1 → v1 progression
- **Cluster vs namespace scoping**: Right scope for each resource type

---

## 🤔 Reflection Questions

1. **Extension Power**: How do CRDs make Kubernetes infinitely extensible?
2. **API Design**: Why does ODH use multiple API groups instead of one?
3. **Schema Evolution**: How might a CRD evolve from v1alpha1 to v1?
4. **Scope Decisions**: Why is DataScienceCluster cluster-scoped but components could be namespaced?
5. **Controller Contract**: What's the relationship between CRD, custom resource, and controller?

---

## 📈 Connection to Bigger Picture

### What You've Learned:
- **Day 1**: How Kubernetes APIs work fundamentally
- **Day 2**: How to extend those APIs with custom types

### What's Coming:
- **Day 3**: How controllers watch and react to your custom resources
- **Day 4**: Go programming patterns that make it all work
- **Day 8+**: How Kubebuilder generates all this code for you

### Real-World Impact:
You now understand how ODH (and any Kubernetes operator) extends the platform. Every custom resource you saw today started as Go code that Kubebuilder turned into CRDs that controllers watch and manage.

---

## ⏰ Time Check

- **Study Topics**: 30 minutes (target: 40)
- **Hands-on Exercises**: 5 minutes (target: 20)
- **Total**: 35 minutes (target: 60)

---

## ✅ Ready for Day 3?

You should now understand:
- ✅ What CRDs are and why they're essential
- ✅ The three-part structure (metadata, spec, status)
- ✅ How OpenAPI v3 schemas provide validation
- ✅ How ODH uses CRDs to create domain-specific APIs
- ✅ The relationship between Go types and CRD definitions

**Next up**: Day 3 will explore how controllers watch your custom resources and implement the reconciliation logic that makes them actually work!

---

## 📝 Notes Section

**My Key Insights:**
  Learned the building block of CRDs like kubebuilder and how its used to make CRD yamls

**Questions for Later:**


**Cool Discoveries:**
  The use of OpenAPI for API validation

**Connections to Day 1:**
  How we can extend the existing API of k8s


