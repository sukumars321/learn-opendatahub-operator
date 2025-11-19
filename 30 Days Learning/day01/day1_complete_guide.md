# Day 1 Complete Study Guide: Kubernetes API Fundamentals

## 🎯 Day 1 Goal
Understand how Kubernetes APIs work and their relationship to operators, using the OpenDataHub Operator as a real-world example.

---

## 📚 Study Topics (45 minutes)

### 1. Kubernetes API Server Architecture (15 minutes)

#### Core Concepts:
- **API Server**: Central hub that processes all REST requests
- **etcd**: Distributed key-value store that holds all cluster state
- **Controllers**: Background processes that watch API changes and ensure desired state
- **Declarative Model**: You declare what you want, controllers make it happen

#### How It Works in ODH:
1. You create a `DataScienceCluster` resource
2. API server validates and stores it in etcd
3. ODH controller notices the change
4. Controller reconciles by creating/updating components

#### Key Points:
- Everything in Kubernetes is an API object (pods, services, even your custom resources)
- API server is stateless - all state lives in etcd
- Controllers implement the "reconciliation loop" pattern
- Changes trigger events that controllers can watch

### 2. REST API Patterns in Kubernetes (15 minutes)

#### HTTP Methods:
- **GET**: Read resources (`kubectl get pods`)
- **POST**: Create new resources (`kubectl create`)
- **PUT**: Replace entire resource (`kubectl replace`)
- **PATCH**: Update part of resource (`kubectl patch`)
- **DELETE**: Remove resources (`kubectl delete`)

#### URL Structure:
```
/api/v1/namespaces/{namespace}/pods/{name}
/apis/apps/v1/namespaces/{namespace}/deployments/{name}
/apis/datasciencecluster.opendatahub.io/v1/datascienceclusters/{name}
```

#### Response Codes:
- **200**: Success (GET, PATCH)
- **201**: Created (POST)
- **404**: Not Found
- **409**: Conflict (resource already exists)
- **422**: Validation Error

### 3. API Groups, Versions, and Resources (15 minutes)

#### Core Group (`/api/v1`):
- pods, services, configmaps, secrets, namespaces
- Most fundamental Kubernetes objects
- Been stable since Kubernetes 1.0

#### Named Groups (`/apis/{group}/{version}`):
- **apps/v1**: deployments, replicasets, daemonsets
- **batch/v1**: jobs, cronjobs
- **rbac.authorization.k8s.io/v1**: roles, rolebindings
- **datasciencecluster.opendatahub.io/v1**: ODH's custom resources

#### Version Progression:
- **v1alpha1**: Experimental, may be removed
- **v1beta1**: Pre-release, API may change
- **v1**: Stable, backward compatible

#### ODH API Groups:
- `datasciencecluster.opendatahub.io/v1`
- `dscinitialization.opendatahub.io/v1`
- `components.platform.opendatahub.io/v1alpha1`
- `services.platform.opendatahub.io/v1alpha1`

---

## 🔬 Hands-on Exercises (15 minutes)

### Exercise 1: Understand ODH CRD Structure

Look at the main ODH CRD definition we found:

```yaml
# From: /config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: datascienceclusters.datasciencecluster.opendatahub.io
spec:
  group: datasciencecluster.opendatahub.io    # API group
  names:
    kind: DataScienceCluster                  # Object type in Go/YAML
    plural: datascienceclusters               # URL path component
    shortNames: [dsc]                         # kubectl shorthand
    singular: datasciencecluster              # Human-readable name
  scope: Cluster                              # Not namespaced
  versions:
  - name: v1                                  # API version
```

**Analysis Questions:**
1. What's the full API path for this resource?
   - Answer: `/apis/datasciencecluster.opendatahub.io/v1/datascienceclusters`

2. What kubectl commands work with this?
   - `kubectl get dsc` (using shortname)
   - `kubectl get datascienceclusters`
   - `kubectl describe datasciencecluster my-dsc`

3. Is this namespaced or cluster-scoped?
   - Cluster-scoped (scope: Cluster)

### Exercise 2: Explore API Object Structure

Every Kubernetes object has this structure:

```yaml
apiVersion: datasciencecluster.opendatahub.io/v1
kind: DataScienceCluster
metadata:                           # Object identity and metadata
  name: default-dsc
  labels:
    app.kubernetes.io/name: odh
spec:                              # Desired state (what you want)
  components:
    dashboard:
      managementState: Managed
    workbenches:
      managementState: Managed
status:                            # Current state (what actually exists)
  phase: Ready
  conditions:
  - type: Ready
    status: "True"
    reason: ReconcileCompleted
```

**Key Observations:**
- `metadata`: Who/what/where this object is
- `spec`: What you want to happen (input)
- `status`: What's actually happening (output)
- Controllers read `spec`, update `status`

### Exercise 3: Compare Core vs Custom Resources

**Core Kubernetes Pod:**
```yaml
apiVersion: v1                     # Core API
kind: Pod
metadata:
  name: my-pod
  namespace: default               # Namespaced
spec:
  containers: [...]
```

**ODH Custom Resource:**
```yaml
apiVersion: datasciencecluster.opendatahub.io/v1  # Custom API
kind: DataScienceCluster
metadata:
  name: my-dsc                     # No namespace (cluster-scoped)
spec:
  components: [...]
```

**Similarities:**
- Same basic structure (apiVersion, kind, metadata, spec, status)
- Same REST API patterns
- Same kubectl commands work
- Both trigger controller reconciliation

**Differences:**
- Different API groups and versions
- Different scoping (namespace vs cluster)
- Different fields in spec/status
- Different controllers handle them

### Exercise 4: Trace API Request Flow

When you run `kubectl apply -f datasciencecluster.yaml`:

1. **kubectl** reads YAML, converts to JSON
2. **API Server** receives POST request to `/apis/datasciencecluster.opendatahub.io/v1/datascienceclusters`
3. **Authentication**: Are you who you say you are?
4. **Authorization**: Are you allowed to create this resource?
5. **Admission Control**: Any webhooks that want to validate/modify?
6. **Validation**: Does the resource match the OpenAPI schema?
7. **Storage**: Resource is stored in etcd
8. **Watches**: Controllers get notified of the change
9. **Reconciliation**: ODH controller creates/updates actual components

---

## 🔍 ODH Code References

### Main CRD Files to Explore:
```bash
# Primary ODH resources
/config/crd/bases/datasciencecluster.opendatahub.io_datascienceclusters.yaml
/config/crd/bases/dscinitialization.opendatahub.io_dscinitializations.yaml

# Component resources
/config/crd/bases/components.platform.opendatahub.io_dashboards.yaml
/config/crd/bases/components.platform.opendatahub.io_workbenches.yaml
/config/crd/bases/components.platform.opendatahub.io_modelmeshservings.yaml
```

### Type Definitions:
```bash
# Go struct definitions (what we'll study on Day 4)
/apis/datasciencecluster/v1/datasciencecluster_types.go
/apis/dscinitialization/v1/dscinitialization_types.go
```

---

## 🧠 Key Takeaways

1. **Everything is an API Object**: Pods, services, and ODH resources all follow the same patterns
2. **Controllers Drive Reconciliation**: They watch for changes and make them happen
3. **CRDs Extend the API**: ODH adds new resource types without modifying Kubernetes core
4. **REST Patterns Apply**: Standard HTTP methods and status codes work everywhere
5. **API Groups Organize Functionality**: Related resources are grouped together

## 🤔 Reflection Questions

1. How does the DataScienceCluster CRD extend Kubernetes' capabilities?
2. Why might ODH use cluster-scoped resources instead of namespaced ones?
3. What happens when you create a DataScienceCluster resource?
4. How do API groups help organize different types of functionality?

## ⏰ Time Check

- **Study Topics**: ____ minutes (target: 45)
- **Hands-on Exercises**: ____ minutes (target: 15)
- **Total**: ____ minutes (target: 60)

## ✅ Ready for Day 2?

You should now understand:
- ✅ How Kubernetes API server processes requests
- ✅ REST API patterns in Kubernetes
- ✅ API groups, versions, and resources
- ✅ How ODH extends Kubernetes with custom resources
- ✅ Basic structure of API objects (metadata, spec, status)

**Next up**: Day 2 will dive deeper into Custom Resource Definitions and how they work!

---

## 📝 Notes Section

Use this space to jot down:
- Questions that came up
- Interesting discoveries
- Connections to your existing knowledge
- Things you want to explore further

**My Notes:**
[Write your thoughts here]

**Questions for later:**
[Any confusion or curiosity to follow up on]