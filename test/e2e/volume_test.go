package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/sapcc/kubernikus/pkg/util/generator"
	"github.com/sapcc/kubernikus/test/e2e/framework"
)

const (
	TestWaitForPVCBoundTimeout = 5 * time.Minute
	TestWaitForPVCPodsRunning  = 15 * time.Minute
)

type VolumeTests struct {
	Kubernetes *framework.Kubernetes
	Namespace  string
	Nodes      *v1.NodeList
}

func (v *VolumeTests) Run(t *testing.T) {
	runParallel(t)

	v.Namespace = generator.SimpleNameGenerator.GenerateName("e2e-volumes-")

	var err error
	v.Nodes, err = v.Kubernetes.ClientSet.CoreV1().Nodes().List(context.Background(), meta_v1.ListOptions{})
	require.NoError(t, err, "There must be no error while listing the kluster's nodes")
	require.NotEmpty(t, v.Nodes.Items, "No nodes returned by list")

	//defer t.Run("Cleanup", v.DeleteNamespace)
	t.Run("CreateNamespace", v.CreateNamespace)
	t.Run("WaitNamespace", v.WaitForNamespace)
	t.Run("CreatePVC", v.CreatePVC)
	t.Run("CreatePod", v.CreatePod)
	t.Run("WaitPVCBound", v.WaitForPVCBound)
	t.Run("WaitPodRunning", v.WaitForPVCPodsRunning)
}

func (p *VolumeTests) CreateNamespace(t *testing.T) {
	_, err := p.Kubernetes.ClientSet.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{ObjectMeta: meta_v1.ObjectMeta{Name: p.Namespace}}, meta_v1.CreateOptions{})
	require.NoError(t, err, "There must be no error while creating a namespace")
}

func (p *VolumeTests) WaitForNamespace(t *testing.T) {
	err := p.Kubernetes.WaitForDefaultServiceAccountInNamespace(p.Namespace)
	require.NoError(t, err, "There must be no error while waiting for the namespace")
}

func (p *VolumeTests) DeleteNamespace(t *testing.T) {
	err := p.Kubernetes.ClientSet.CoreV1().Namespaces().Delete(context.Background(), p.Namespace, meta_v1.DeleteOptions{})
	require.NoError(t, err, "There must be no error while deleting a namespace")
}

func (p *VolumeTests) CreatePod(t *testing.T) {
	_, err := p.Kubernetes.ClientSet.CoreV1().Pods(p.Namespace).Create(context.Background(), &v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "pvc-hostname",
			Namespace: p.Namespace,
			Labels: map[string]string{
				"app": "pvc-hostname",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: ServeHostnameImage,
					Name:  "pvc-hostname",
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "pvc-hostname",
							MountPath: "/mymount",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "pvc-hostname",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc-hostname",
						},
					},
				},
			},
		},
	}, meta_v1.CreateOptions{})
	assert.NoError(t, err, "There should be no error while creating a pod with a volume")
}

func (p *VolumeTests) WaitForPVCPodsRunning(t *testing.T) {
	label := labels.SelectorFromSet(labels.Set(map[string]string{"app": "pvc-hostname"}))
	_, err := p.Kubernetes.WaitForPodsWithLabelRunningReady(p.Namespace, label, 1, TestWaitForPVCPodsRunning)
	require.NoError(t, err, "There must be no error while waiting for the pod with mounted volume to become ready")
}

func (p *VolumeTests) CreatePVC(t *testing.T) {
	_, err := p.Kubernetes.ClientSet.CoreV1().PersistentVolumeClaims(p.Namespace).Create(context.Background(), &v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Namespace: p.Namespace,
			Name:      "pvc-hostname",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}, meta_v1.CreateOptions{})
	assert.NoError(t, err)
}

func (p *VolumeTests) WaitForPVCBound(t *testing.T) {
	err := p.Kubernetes.WaitForPVCBound(p.Namespace, "pvc-hostname", TestWaitForPVCBoundTimeout)
	require.NoError(t, err, "There must be no error while waiting for the PVC to be bound")
}
