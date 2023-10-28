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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type InotifyFs struct {
	MaxQueuedEvents  int `json:"max_queued_events,omitempty"`
	MaxUserInstances int `json:"max_user_instances,omitempty"`
	MaxUserWatches   int `json:"max_user_watches,omitempty"`
}

type CoreNet struct {
	RmemDefault int `json:"rmem_default,omitempty"`
	RmemMax     int `json:"rmem_max,omitempty"`
	WmemDefault int `json:"wmem_default,omitempty"`
	WmemMax     int `json:"wmem_max,omitempty"`
}

type Ipv4Net struct {
	TcpMem                []int `json:"tcp_mem,omitempty"`
	TcpRmem               []int `json:"tcp_rmem,omitempty"`
	TcpSlowStartAfterIdle int   `json:"tcp_slow_start_after_idle,omitempty"`
	TcpWmem               []int `json:"tcp_wmem,omitempty"`
	UdpMem                []int `json:"udp_mem,omitempty"`
	UdpRmemMin            int   `json:"udp_rmem_min,omitempty"`
	UdpWmemMin            int   `json:"udp_wmem_min,omitempty"`
}

type NetfilterNet struct {
	NfConntrackTcpTimeoutMaxRetrans     int `json:"nf_conntrack_tcp_timeout_max_retrans,omitempty"`
	NfConntrackTcpTimeoutUnacknowledged int `json:"nf_conntrack_tcp_timeout_unacknowledged,omitempty"`
}

type SctpNet struct {
	AuthEnable int   `json:"auth_enable,omitempty"`
	SctpMem    []int `json:"sctp_mem,omitempty"`
}

type FsSysctl struct {
	Inotify InotifyFs `json:"inotify,omitempty"`
}

type KernelSysctl struct {
	Core_pattern        string `json:"core_pattern,omitempty"`
	Sched_rt_runtime_us int    `json:"sched_rt_runtime_us,omitempty"`
}

type NetSysctl struct {
	Core      CoreNet      `json:"core,omitempty"`
	Ipv4      Ipv4Net      `json:"ipv4,omitempty"`
	Netfilter NetfilterNet `json:"netfilter,omitempty"`
	Sctp      SctpNet      `json:"sctp,omitempty"`
}

type VmSysctl struct {
	MaxMapCount int `json:"max_map_count,omitempty"`
}

type Sysctl struct {
	Fs     FsSysctl     `json:"fs,omitempty"`
	Kernel KernelSysctl `json:"kernel,omitempty"`
	Net    NetSysctl    `json:"net,omitempty"`
	Vm     VmSysctl     `json:"vm,omitempty"`
}

// NodeSystemConfigUpdateSpec defines the desired state of NodeSystemConfigUpdate
type NodeSystemConfigUpdateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NodeSystemConfigUpdate. Edit nodesystemconfigupdate_types.go to remove/update
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Sysctl       Sysctl            `json:"sysctl,omitempty"`
}

// NodeSystemConfigUpdateStatus defines the observed state of NodeSystemConfigUpdate
type NodeSystemConfigUpdateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// NodeSystemConfigUpdate is the Schema for the nodesystemconfigupdates API
type NodeSystemConfigUpdate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSystemConfigUpdateSpec   `json:"spec,omitempty"`
	Status NodeSystemConfigUpdateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeSystemConfigUpdateList contains a list of NodeSystemConfigUpdate
type NodeSystemConfigUpdateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeSystemConfigUpdate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeSystemConfigUpdate{}, &NodeSystemConfigUpdateList{})
}
