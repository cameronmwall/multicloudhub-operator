package helmrepo

import (
	"reflect"
	"testing"

	operatorsv1beta1 "github.com/open-cluster-management/multicloudhub-operator/pkg/apis/operators/v1beta1"
	"github.com/open-cluster-management/multicloudhub-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeployment(t *testing.T) {
	replicas := int(1)
	empty := &operatorsv1beta1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: operatorsv1beta1.MultiClusterHubSpec{
			Version:         "",
			ImageRepository: "",
			ImagePullPolicy: "",
			ImagePullSecret: "",
			ImageTagSuffix:  "",
			Mongo:           operatorsv1beta1.Mongo{},
			ReplicaCount:    &replicas,
		},
	}

	cs := utils.CacheSpec{
		IngressDomain:   "",
		ImageShaDigests: map[string]string{},
	}

	t.Run("MCH with empty fields", func(t *testing.T) {
		_ = Deployment(empty, cs)
	})

	essentialsOnly := &operatorsv1beta1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: operatorsv1beta1.MultiClusterHubSpec{
			Version:         "test",
			ImageRepository: "test",
			ImagePullPolicy: "test",
			ImageTagSuffix:  "test",
			ReplicaCount:    &replicas,
		},
	}
	t.Run("MCH with only required values", func(t *testing.T) {
		_ = Deployment(essentialsOnly, cs)
	})
}

func TestService(t *testing.T) {
	mch := &operatorsv1beta1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testName",
			Namespace: "testNS",
		},
	}

	t.Run("Create service", func(t *testing.T) {
		s := Service(mch)
		if ns := s.Namespace; ns != "testNS" {
			t.Errorf("expected namespace %s, got %s", "testNS", ns)
		}
		if ref := s.GetOwnerReferences(); ref[0].Name != "testName" {
			t.Errorf("expected ownerReference %s, got %s", "testName", ref[0].Name)
		}
	})
}

func TestValidateDeployment(t *testing.T) {
	replicas := int(1)
	mch := &operatorsv1beta1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: operatorsv1beta1.MultiClusterHubSpec{
			Version:         "1.0.0",
			ImageRepository: "quay.io/open-cluster-management",
			ImagePullPolicy: "Always",
			ImagePullSecret: "test",
			ReplicaCount:    &replicas,
			Mongo:           operatorsv1beta1.Mongo{},
			NodeSelector: map[string]string{
				"test": "test",
			},
		},
	}

	cs := utils.CacheSpec{
		IngressDomain:   "testIngress",
		ImageShaDigests: map[string]string{},
	}

	// 1. Valid mch
	dep := Deployment(mch, cs)

	// 2. Modified ImagePullSecret
	dep1 := dep.DeepCopy()
	dep1.Spec.Template.Spec.ImagePullSecrets = nil

	// 3. Modified image
	dep2 := dep.DeepCopy()
	dep2.Spec.Template.Spec.Containers[0].Image = "differentImage"

	// 4. Modified pullPolicy
	dep3 := dep.DeepCopy()
	dep3.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullNever

	// 5. Modified NodeSelector
	dep4 := dep.DeepCopy()
	dep4.Spec.Template.Spec.NodeSelector = nil

	// 6. Modified ReplicaCount
	dep5 := dep.DeepCopy()
	dep5.Spec.Replicas = new(int32)

	type args struct {
		m   *operatorsv1beta1.MultiClusterHub
		dep *appsv1.Deployment
	}
	tests := []struct {
		name  string
		args  args
		want  *appsv1.Deployment
		want1 bool
	}{
		{
			name:  "Valid Deployment",
			args:  args{mch, dep},
			want:  dep,
			want1: false,
		},
		{
			name:  "Modified ImagePullSecret",
			args:  args{mch, dep1},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified Image",
			args:  args{mch, dep2},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified PullPolicy",
			args:  args{mch, dep3},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified NodeSelector",
			args:  args{mch, dep4},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified ReplicaCount",
			args:  args{mch, dep5},
			want:  dep,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ValidateDeployment(tt.args.m, cs, tt.args.dep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateDeployment() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValidateDeployment() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}