package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nginxv1 "example.com/nginx-controller/api/v1"
)

// +kubebuilder:docs-gen:collapse=Imports

const (
	TestNginxName      = "test-nginx"
	TestNamespace      = "test"
	TestReplica        = int32(3)
	TestDeploymentName = "deploy-" + TestNginxName
	TestServiceName    = "service-" + TestNginxName
)

var _ = Describe("nginx controller", func() {

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &nginxv1.Nginx{}, client.InNamespace(TestNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace(TestNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.Service{}, client.InNamespace(TestNamespace))
		Expect(err).NotTo(HaveOccurred())

	})

	Context("When updating Nginx", func() {

		replicas := TestReplica

		// Nginxを作成したらDeploymentとServiceが作成されることの確認
		It("Should create Deployment and Service", func() {

			// Nginxを作成
			By("By creating a new Nginx")
			nginx := newNginx(&replicas)
			err := k8sClient.Create(ctx, nginx)
			Expect(err).NotTo(HaveOccurred())

			// Nginxの作成に伴って作成されるDeploymentを取得
			By("By checking the Deployment")
			deploy := appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: TestNamespace, Name: TestDeploymentName}, &deploy)
			}).Should(Succeed())

			By("By checking the Deployment has same replicas with Nginx")
			Expect(deploy.Spec.Replicas).Should(Equal(&replicas))

			By("By checking the Service")
			service := corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: TestNamespace, Name: TestServiceName}, &service)
			}).Should(Succeed())

			// Selectorが正しく設定されていることを確認
			By("By checking the Service has Selector correspond to Nginx")
			Expect(service.Spec.Selector).Should(Equal(map[string]string{"controller": TestNginxName}))
			// Portが正しく設定されていることを確認
			By("By checking the Service has Port setting for Nginx")
			Expect(service.Spec.Ports).Should(Equal(
				[]corev1.ServicePort{
					{
						Protocol:   corev1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 80},
					},
				}))

		})

		// Nginxを更新したらDeploymentが更新されることの確認
		It("Should update Deployment", func() {
			// Nginxを作成
			By("By creating a new Nginx")
			nginx := newNginx(&replicas)
			err := k8sClient.Create(ctx, nginx)
			Expect(err).NotTo(HaveOccurred())

			// Nginxを更新
			By("By updating the previous Nginx")
			newReplica := int32(1)
			nginx.Spec.Replicas = &newReplica
			err = k8sClient.Update(ctx, nginx, &client.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond) // Nginx ControllerがDeploymentのReplicasを更新するまで待機

			// Nginxの作成に伴って作成されるDeploymentを取得
			By("By checking the Deployment has same replicas with Nginx")
			deploy := appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: TestNamespace, Name: TestDeploymentName}, &deploy)
			}).Should(Succeed())

			time.Sleep(100 * time.Millisecond)

			Expect(deploy.Spec.Replicas).Should(Equal(&newReplica))

			// NginxのStatus.AvailableReplicasをチェックしたいけどtestenvではDeployment配下のPodが作成されないので判定不可
			// time.Sleep(100 * time.Millisecond)

			// // Nginxを取得
			// By("By checking the Nginx Status")
			// updatedNginx := nginxv1.Nginx{}
			// Eventually(func() error {
			// 	return k8sClient.Get(ctx, client.ObjectKey{Namespace: TestNamespace, Name: TestNginxName}, &updatedNginx)
			// }).Should(Succeed())

			// Expect(updatedNginx.Status.AvailableReplicas).Should(Equal(&newReplica))

		})

	})

})

// Nginxオブジェクトを生成する関数
func newNginx(replicas *int32) *nginxv1.Nginx {

	return &nginxv1.Nginx{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TestNginxName,
			Namespace: TestNamespace,
		},
		Spec: nginxv1.NginxSpec{
			Replicas: replicas,
		},
	}
}
