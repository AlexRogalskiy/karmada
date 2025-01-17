package core

import (
	"testing"

	workv1alpha2 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha2"
	utilhelper "github.com/karmada-io/karmada/pkg/util/helper"
	"github.com/karmada-io/karmada/test/helper"
)

const (
	ClusterMember1 = "member1"
	ClusterMember2 = "member2"
	ClusterMember3 = "member3"
	ClusterMember4 = "member4"
)

func Test_dispenser_takeByWeight(t *testing.T) {
	tests := []struct {
		name        string
		numReplicas int32
		result      []workv1alpha2.TargetCluster
		weightList  utilhelper.ClusterWeightInfoList
		desired     []workv1alpha2.TargetCluster
		done        bool
	}{
		{
			name:        "Scale up 6 replicas",
			numReplicas: 6,
			result: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 1},
				{Name: "B", Replicas: 2},
				{Name: "C", Replicas: 3},
			},
			weightList: []utilhelper.ClusterWeightInfo{
				{ClusterName: "A", Weight: 1},
				{ClusterName: "B", Weight: 2},
				{ClusterName: "C", Weight: 3},
			},
			desired: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 2},
				{Name: "B", Replicas: 4},
				{Name: "C", Replicas: 6},
			},
			done: true,
		},
		{
			name:        "Scale up 3 replicas",
			numReplicas: 3,
			result: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 1},
				{Name: "B", Replicas: 2},
				{Name: "C", Replicas: 3},
			},
			weightList: []utilhelper.ClusterWeightInfo{
				{ClusterName: "A", Weight: 1},
				{ClusterName: "B", Weight: 2},
				{ClusterName: "C", Weight: 3},
			},
			desired: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 1},
				{Name: "B", Replicas: 3},
				{Name: "C", Replicas: 5},
			},
			done: true,
		},
		{
			name:        "Scale up 2 replicas",
			numReplicas: 2,
			result: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 1},
				{Name: "B", Replicas: 2},
				{Name: "C", Replicas: 3},
			},
			weightList: []utilhelper.ClusterWeightInfo{
				{ClusterName: "A", Weight: 1},
				{ClusterName: "B", Weight: 2},
				{ClusterName: "C", Weight: 3},
			},
			desired: []workv1alpha2.TargetCluster{
				{Name: "A", Replicas: 1},
				{Name: "B", Replicas: 2},
				{Name: "C", Replicas: 5},
			},
			done: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newDispenser(tt.numReplicas, tt.result)
			a.takeByWeight(tt.weightList)
			if a.done() != tt.done {
				t.Errorf("expected after takeByWeight: %v, but got: %v", tt.done, a.done())
			}
			if !helper.IsScheduleResultEqual(a.result, tt.desired) {
				t.Errorf("expected result after takeByWeight: %v, but got: %v", tt.desired, a.result)
			}
		})
	}
}

func Test_dynamicDivideReplicas(t *testing.T) {
	tests := []struct {
		name    string
		state   *assignState
		want    []workv1alpha2.TargetCluster
		wantErr bool
	}{
		{
			name: "replica 12, dynamic weight 18:12:6",
			state: &assignState{
				availableClusters: TargetClustersList{
					workv1alpha2.TargetCluster{Name: ClusterMember1, Replicas: 18},
					workv1alpha2.TargetCluster{Name: ClusterMember2, Replicas: 12},
					workv1alpha2.TargetCluster{Name: ClusterMember3, Replicas: 6},
				},
				targetReplicas:    12,
				availableReplicas: 36,
				strategyType:      DynamicWeightStrategy,
			},
			want: []workv1alpha2.TargetCluster{
				{Name: ClusterMember1, Replicas: 6},
				{Name: ClusterMember2, Replicas: 4},
				{Name: ClusterMember3, Replicas: 2},
			},
			wantErr: false,
		},
		{
			name: "replica 12, dynamic weight 20:12:6",
			state: &assignState{
				availableClusters: TargetClustersList{
					workv1alpha2.TargetCluster{Name: ClusterMember1, Replicas: 20},
					workv1alpha2.TargetCluster{Name: ClusterMember2, Replicas: 12},
					workv1alpha2.TargetCluster{Name: ClusterMember3, Replicas: 6},
				},
				targetReplicas:    12,
				availableReplicas: 38,
				strategyType:      DynamicWeightStrategy,
			},
			want: []workv1alpha2.TargetCluster{
				{Name: ClusterMember1, Replicas: 7},
				{Name: ClusterMember2, Replicas: 4},
				{Name: ClusterMember3, Replicas: 1},
			},
			wantErr: false,
		},
		{
			name: "replica 12, dynamic weight 6:12:6",
			state: &assignState{
				availableClusters: TargetClustersList{
					workv1alpha2.TargetCluster{Name: ClusterMember1, Replicas: 6},
					workv1alpha2.TargetCluster{Name: ClusterMember2, Replicas: 12},
					workv1alpha2.TargetCluster{Name: ClusterMember3, Replicas: 6},
				},
				targetReplicas:    12,
				availableReplicas: 24,
				strategyType:      DynamicWeightStrategy,
			},
			want: []workv1alpha2.TargetCluster{
				{Name: ClusterMember1, Replicas: 3},
				{Name: ClusterMember2, Replicas: 6},
				{Name: ClusterMember3, Replicas: 3},
			},
			wantErr: false,
		},
		{
			name: "replica 12, aggregated 12:6:6",
			state: &assignState{
				availableClusters: TargetClustersList{
					workv1alpha2.TargetCluster{Name: ClusterMember2, Replicas: 12},
					workv1alpha2.TargetCluster{Name: ClusterMember1, Replicas: 6},
					workv1alpha2.TargetCluster{Name: ClusterMember3, Replicas: 6},
				},
				targetReplicas:    12,
				availableReplicas: 24,
				strategyType:      AggregatedStrategy,
			},
			want: []workv1alpha2.TargetCluster{
				{Name: ClusterMember2, Replicas: 12},
			},
			wantErr: false,
		},
		{
			name: "replica 12, aggregated 6:6:6",
			state: &assignState{
				availableClusters: TargetClustersList{
					workv1alpha2.TargetCluster{Name: ClusterMember1, Replicas: 6},
					workv1alpha2.TargetCluster{Name: ClusterMember2, Replicas: 6},
					workv1alpha2.TargetCluster{Name: ClusterMember3, Replicas: 6},
				},
				targetReplicas:    12,
				availableReplicas: 18,
				strategyType:      AggregatedStrategy,
			},
			want: []workv1alpha2.TargetCluster{
				{Name: ClusterMember1, Replicas: 6},
				{Name: ClusterMember2, Replicas: 6},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dynamicDivideReplicas(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("dynamicDivideReplicas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !helper.IsScheduleResultEqual(got, tt.want) {
				t.Errorf("dynamicDivideReplicas() got = %v, want %v", got, tt.want)
			}
		})
	}
}
