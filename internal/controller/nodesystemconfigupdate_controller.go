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
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

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

func insertUnderscored(input string) string {
	var result strings.Builder

	for i, char := range input {
		if i > 0 && unicode.IsUpper(char) {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}

	return result.String()
}

func convertToString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, value.Type().Bits())
	case reflect.String:
		return value.String()
	default:
		return fmt.Sprintf("%v", value.Interface())
	}
}

func processSysctlParameters(sysctls systemv1alpha1.Sysctl, basePath string) error {
	err := processSysctlFields(sysctls, basePath)
	if err != nil {
		return err
	}

	return nil
}

func processSysctlFields(field interface{}, basePath string) error {
	value := reflect.ValueOf(field)

	if !value.IsValid() || value.IsZero() {
		return nil
	}

	if value.Kind() == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			subfield := value.Field(i).Interface()
			fieldName := value.Type().Field(i).Name

			fieldNameUnderscore := insertUnderscored(fieldName)

			subpath := strings.Join([]string{basePath, fieldNameUnderscore}, "/")

			if err := processSysctlFields(subfield, subpath); err != nil {
				return err
			}
		}

	} else if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		sliceValues := []string{}
		for i := 0; i < value.Len(); i++ {
			element := value.Index(i)
			elementString := convertToString(element)
			sliceValues = append(sliceValues, elementString)
		}
		sliceValueString := strings.Join(sliceValues, "\t")

		currentValue, err := readValueFromFile(strings.ToLower(basePath))
		if err != nil {
			return err
		}

		if sliceValueString != currentValue {
			err = writeValueToFile(strings.ToLower(basePath), sliceValueString)
			if err != nil {
				return err
			}
		}

	} else {
		//fieldName := basePath[strings.LastIndex(basePath, "/")+1:]
		//filePath := strings.Join([]string{basePath, fieldName}, "/")

		currentValue, err := readValueFromFile(strings.ToLower(basePath))
		if err != nil {
			return err
		}

		desiredValue := fmt.Sprintf("%v", value)
		if desiredValue != currentValue {
			err = writeValueToFile(strings.ToLower(basePath), field)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func readValueFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func writeValueToFile(filePath string, value interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	strValue := fmt.Sprintf("%v", value)
	_, err = file.WriteString(strValue)
	if err != nil {
		return err
	}

	return nil
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

	if err := r.Get(ctx, req.NamespacedName, nodeSysConfUpdate); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.Get(ctx, client.ObjectKey{Name: os.Getenv("NODENAME")}, nodeData); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if mapsMatch(nodeSysConfUpdate.Spec.NodeSelector, nodeData.GetLabels()) {
		err := processSysctlParameters(nodeSysConfUpdate.Spec.Sysctl, "/sysctls")
		if err != nil {
			l.Info("NODE_MATCH", "SELECTOR", nodeSysConfUpdate.Spec.NodeSelector)
			l.Info("NOT_PROCESSED", "CONFIG", nodeSysConfUpdate.ObjectMeta.Name, "ERROR", err)
		} else {
			l.Info("NODE_MATCH", "SELECTOR", nodeSysConfUpdate.Spec.NodeSelector)

			// Update Status fields
			nodeSysConfUpdate.Status.LastUpdateTime = time.Now().Format(time.RFC3339)
			if nodeSysConfUpdate.Status.NodesConfigured == nil {
				nodeSysConfUpdate.Status.NodesConfigured = make(map[string]bool)
			}

			if _, exists := nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")]; !exists {
				// Node not configured, so set it as configured
				nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")] = true
			} else {
				nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")] = true
			}

			if err := r.Status().Update(ctx, nodeSysConfUpdate); err != nil {
				return ctrl.Result{}, err
			}

			l.Info("PROCESSED", "CONFIG", nodeSysConfUpdate.ObjectMeta.Name)
		}
	} else {
		l.Info("NO_NODE_MATCH", "SELECTOR", nodeSysConfUpdate.Spec.NodeSelector)

		// Update Status fields
		nodeSysConfUpdate.Status.LastUpdateTime = time.Now().Format(time.RFC3339)
		if nodeSysConfUpdate.Status.NodesConfigured == nil {
			nodeSysConfUpdate.Status.NodesConfigured = make(map[string]bool)
		}

		if _, exists := nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")]; !exists {
			nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")] = false
		} else {
			nodeSysConfUpdate.Status.NodesConfigured[os.Getenv("NODENAME")] = false
		}

		if err := r.Status().Update(ctx, nodeSysConfUpdate); err != nil {
			return ctrl.Result{}, err
		}

		l.Info("NOT_PROCESSED", "CONFIG", nodeSysConfUpdate.ObjectMeta.Name)
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 15 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeSystemConfigUpdateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&systemv1alpha1.NodeSystemConfigUpdate{}).
		Complete(r)
}
