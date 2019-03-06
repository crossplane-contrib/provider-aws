/*
Copyright 2018 The Crossplane Authors.

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
	"context"
	"log"
	"testing"

	"github.com/go-test/deep"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane/pkg/test"
)

const (
	namespace = "default"
	name      = "test-instance"
)

var (
	cfg *rest.Config
	c   client.Client
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	t := test.NewEnv(namespace, test.CRDs())
	cfg = t.Start()

	if c, err = client.New(cfg, client.Options{Scheme: scheme.Scheme}); err != nil {
		log.Fatal(err)
	}

	t.StopAndExit(m.Run())
}

func TestStorageReplicationGroup(t *testing.T) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	created := &ReplicationGroup{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	g.Expect(c.Create(ctx, created)).NotTo(gomega.HaveOccurred())

	fetched := &ReplicationGroup{}
	g.Expect(c.Get(ctx, key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(ctx, updated)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(ctx, key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(ctx, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(ctx, key, fetched)).To(gomega.HaveOccurred())
}

func TestNewReplicationGroupSpec(t *testing.T) {
	cases := []struct {
		name       string
		properties map[string]string
		want       *ReplicationGroupSpec
	}{
		{
			name: "AllProperties",
			properties: map[string]string{

				// Strings
				"authToken":                  "coolPassword",
				"cacheNodeType":              "m1.supercool",
				"cacheParameterGroupName":    "coolParamGroup",
				"cacheSubnetGroupName":       "coolSubnet",
				"engineVersion":              "5.0.0",
				"notificationTopicArn":       "arn:supercool:notifyme",
				"preferredMaintenanceWindow": "05:00-06:00",
				"snapshotName":               "coolsnapshot",
				"snapshotWindow":             "04:00-05:00",

				// Slices
				"cacheSecurityGroupNames":  "coolGroup,coolerGroup",
				"preferredCacheClusterAzs": "ap-southeast-2,eu-north-1",
				"securityGroupIds":         "supersecureGroup,themostsecureGroup",

				// Bools
				"atRestEncryptionEnabled":  "true",
				"automaticFailoverEnabled": "true",
				"transitEncryptionEnabled": "true",

				// Integers
				"numCacheClusters":       "2",
				"numNodeGroups":          "3",
				"port":                   "6379",
				"replicasPerNodeGroup":   "2",
				"snapshotRetentionLimit": "1",
			},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,

				CacheNodeType:              "m1.supercool",
				CacheParameterGroupName:    "coolParamGroup",
				CacheSubnetGroupName:       "coolSubnet",
				EngineVersion:              "5.0.0",
				NotificationTopicARN:       "arn:supercool:notifyme",
				PreferredMaintenanceWindow: "05:00-06:00",
				SnapshotName:               "coolsnapshot",
				SnapshotWindow:             "04:00-05:00",
				CacheSecurityGroupNames:    []string{"coolGroup", "coolerGroup"},
				PreferredCacheClusterAZs:   []string{"ap-southeast-2", "eu-north-1"},
				SecurityGroupIDs:           []string{"supersecureGroup", "themostsecureGroup"},
				AtRestEncryptionEnabled:    true,
				AutomaticFailoverEnabled:   true,
				TransitEncryptionEnabled:   true,
				NumCacheClusters:           2,
				NumNodeGroups:              3,
				Port:                       6379,
				ReplicasPerNodeGroup:       2,
				SnapshotRetentionLimit:     1,
			},
		},
		{
			name:       "NilProperties",
			properties: nil,
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "UnknownProperties",
			properties: map[string]string{"unknown": "wat"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "AtRestEncryptionEnabledNotABool",
			properties: map[string]string{"atRestEncryptionEnabled": "maybe"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "AutomaticFailoverEnabledNotABool",
			properties: map[string]string{"automaticFailoverEnabled": "maybe"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "TransitEncryptionEnabledNotABool",
			properties: map[string]string{"transitEncryptionEnabled": "maybe"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "NumCacheClustersNotANumber",
			properties: map[string]string{"numCacheClusters": "wat"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "NumNodeGroupsNotANumber",
			properties: map[string]string{"numNodeGroups": "wat"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "PortNotANumber",
			properties: map[string]string{"port": "wat"},
			want: &ReplicationGroupSpec{
				ReclaimPolicy: corev1alpha1.ReclaimRetain,
			},
		},
		{
			name:       "CacheSecurityGroupNamesExtraneousWhitespace",
			properties: map[string]string{"cacheSecurityGroupNames": "   value,   suchvalue   "},
			want: &ReplicationGroupSpec{
				ReclaimPolicy:           corev1alpha1.ReclaimRetain,
				CacheSecurityGroupNames: []string{"value", "suchvalue"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewReplicationGroupSpec(tc.properties)
			if diff := deep.Equal(got, tc.want); diff != nil {
				t.Errorf("got != want:\n%v", diff)
			}
		})
	}
}
