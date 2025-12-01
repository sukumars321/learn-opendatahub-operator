# Day 8: Introduction to Kubebuilder

## 🎯 Goal
Understand Kubebuilder as the framework for building Kubernetes operators, focusing on how it simplifies operator development and how ODH uses it.

## 📚 Study Materials

### Core Study Guide
- **`day8_complete_guide.md`** - Complete 45-minute study guide covering:
  - What is Kubebuilder and why use it
  - Kubebuilder architecture and components
  - How ODH operator uses Kubebuilder
  - Project structure and scaffolding
  - Controllers, managers, and reconcilers in Kubebuilder

### Hands-on Exercises
- **`day8_exercises.md`** - Kubebuilder exploration exercises (15 minutes)
- Focus on understanding ODH's Kubebuilder setup and structure

## ⏰ Time Allocation
- **Study**: 45 minutes
- **Code Analysis**: 15 minutes
- **Total**: 60 minutes

## 🎓 Learning Outcomes

By the end of Day 8, you'll understand:
- What Kubebuilder is and its role in operator development
- How Kubebuilder simplifies controller creation
- ODH operator's Kubebuilder project structure
- The relationship between Kubebuilder, controller-runtime, and client-go
- How Kubebuilder generates code and manifests
- Manager pattern and controller registration in Kubebuilder

## 🔗 Prerequisites
- Days 1-7: Kubernetes APIs, CRDs, controllers, Go programming, client-go, controller-runtime

## ➡️ Next Steps
Day 9: Kubebuilder Markers and Code Generation

## 📖 Key Concepts Preview
- **Kubebuilder**: Framework for building Kubernetes operators
- **Scaffolding**: Auto-generating boilerplate code and project structure
- **Markers**: Comments that drive code generation
- **Manager**: Central component that runs controllers
- **Project Layout**: Standard structure for operator projects