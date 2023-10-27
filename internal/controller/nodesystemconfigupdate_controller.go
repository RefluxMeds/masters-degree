/*
Copyright 2023.

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
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	systemv1alpha1 "github.com/RefluxMeds/masters-degree/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Custom functions written
func mapsMatch(mapSelector, mapNode map[string]string) bool {
	for key, valueSelector := range mapSelector {
		valueNode, exists := mapNode[key]

		if !exists || valueSelector != valueNode {
			return false
		}
	}

	return true
}

// NodeSystemConfigUpdateReconciler reconciles a NodeSystemConfigUpdate object
type NodeSystemConfigUpdateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=system.masters.degree,resources=nodesystemconfigupdates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=system.masters.degree,resources=nodesystemconfigupdates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=system.masters.degree,resources=nodesystemconfigupdates/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeSystemConfigUpdate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *NodeSystemConfigUpdateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	nodeSysConfUpdate := &systemv1alpha1.NodeSystemConfigUpdate{}
	nodeData := &corev1.Node{}
	nodeName := os.Getenv("NODENAME")

	if err := r.Get(ctx, req.NamespacedName, nodeSysConfUpdate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.Get(ctx, client.ObjectKey{Name: nodeName}, nodeData); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nodeSelector := nodeSysConfUpdate.Spec.NodeSelector
	nodeLabels := nodeData.GetLabels()

	if mapsMatch(nodeSelector, nodeLabels) {
		l.Info("MatchedNodeLabels", "SelectorLabel", nodeSelector, "NodeLabel", nodeLabels)
	} else {
		l.Info("CheckingNodeLabels", "SelectorLabel", nodeSelector, "NodeLabel", nodeLabels)
	}

	//if err := r.Update(ctx, nodeSysConfUpdate)

	return ctrl.Result{Requeue: true, RequeueAfter: 15 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeSystemConfigUpdateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&systemv1alpha1.NodeSystemConfigUpdate{}).
		Complete(r)
}
