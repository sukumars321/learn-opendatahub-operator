# Day 9: Kubebuilder Markers and Code Generation

## 🎯 Goal
Understand how ODH uses code generation extensively through Kubebuilder markers and automated tooling to maintain consistency and reduce boilerplate code.

## 📚 Study Materials

### Core Study Guide
- **`day9_complete_guide.md`** - Complete 40-minute study guide covering:
  - Understanding Kubebuilder markers and their purpose
  - RBAC markers for permission generation
  - CRD generation markers and validation
  - Webhook markers for admission control
  - Code generation workflow and `make generate`

### Hands-on Exercises
- **`day9_exercises.md`** - ODH codebase exploration exercises (20 minutes)
- Focus on finding and understanding markers throughout the ODH codebase
- Practice running code generation and observing changes

## ⏰ Time Allocation
- **Study**: 40 minutes
- **Code Exploration**: 20 minutes
- **Total**: 60 minutes

## 🎓 Learning Outcomes

By the end of Day 9, you'll understand:
- What Kubebuilder markers are and why they're used
- How RBAC markers generate Kubernetes permissions
- CRD generation markers and OpenAPI schema validation
- Webhook markers for admission controllers
- ODH's extensive use of code generation
- How to run `make generate` and interpret the results
- The relationship between markers and generated manifests

## 🔗 Prerequisites
- Day 8: Introduction to Kubebuilder
- Understanding of CRDs, RBAC, and Kubernetes manifests

## ➡️ Next Steps
Day 10: ODH Controller Architecture Deep Dive

## 📖 Key Concepts Preview
- **Markers**: Special comments that drive code generation (`//+kubebuilder:`)
- **RBAC Markers**: Generate ClusterRole and Role manifests
- **CRD Markers**: Control CRD generation and validation rules
- **Webhook Markers**: Configure admission webhooks
- **Code Generation**: Automated creation of boilerplate code and manifests
- **Make Targets**: Build system integration for generation workflows