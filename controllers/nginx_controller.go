/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"strconv"

	nginxv1 "example.com/nginx-controller/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	OwnerKey = ".metadata.controller"
	apiGVStr = nginxv1.GroupVersion.String()
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Nginxリソースに対応したDeploymentを作成/更新
func (r *NginxReconciler) CreateOrUpdateDeployment(ctx context.Context, log logr.Logger, nginx *nginxv1.Nginx, deploymentName string) error {

	log.Info("CreateOrUpdate Deployment for " + nginx.Name)

	var operationResult controllerutil.OperationResult
	// Deploymentを作成(structの初期化)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: nginx.Namespace,
		},
	}

	// コールバック関数の中でDeployment(deploy)を定義しCreateOrUpdateで作成/更新
	operationResult, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		// コールバック関数funcの中でDeploymentの作成を実施
		// この関数の中で作成したオブジェクトをもとに差分比較を行うらしい
		// https://github.com/kubernetes-sigs/controller-runtime/blob/d242fe21e646f034995c4c93e9bba388a0fdaab9/pkg/controller/controllerutil/controllerutil.go#L210-L217

		// LabelをMapで定義
		labels := map[string]string{
			"app":        "nginx",
			"controller": nginx.Name,
		}

		deploy.ObjectMeta.Labels = labels
		replicas := int32(1) // 初期値
		if nginx.Spec.Replicas != nil {
			replicas = *nginx.Spec.Replicas // Nginx ObjectのSpecからReplicasを取得
		}
		deploy.Spec.Replicas = &replicas // DeploymentにReplicasを設定

		// DeploymentのLabelSelectorにlabelsを設定
		// https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector
		if deploy.Spec.Selector == nil {
			deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		}

		// Pod Templateにlabelsを設定
		// https://pkg.go.dev/k8s.io/api@v0.25.0/core/v1#PodTemplateSpec
		if deploy.Spec.Template.Labels == nil {
			deploy.Spec.Template.Labels = labels
		}

		// Containerをarrayで定義
		// https://pkg.go.dev/k8s.io/api@v0.25.0/core/v1#Container
		containers := []corev1.Container{
			{
				Name:  "nginx",
				Image: "nginx:latest",
			},
		}

		// Pod TemplateにContainerを設定
		if deploy.Spec.Template.Spec.Containers == nil {
			deploy.Spec.Template.Spec.Containers = containers
		}

		// ★DeploymentにOwnerReferenceを設定
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#SetControllerReference
		if err := ctrl.SetControllerReference(nginx, deploy, r.Scheme); err != nil {
			log.Error(err, "Unable to set OwnerReference from Nginx to Deployment")
		}

		return nil

	})
	if err != nil {
		log.Error(err, "Unable to ensure deployment is correct")
		return err
	}

	log.Info("CreateOrUpdate Deployment for " + nginx.Name + ": " + string(operationResult))

	return nil

}

// Nginxリソースに対応したServiceを作成/更新
func (r *NginxReconciler) CreateOrUpdateService(ctx context.Context, log logr.Logger, nginx *nginxv1.Nginx, serviceName string) error {
	log.Info("CreateOrUpdate Service for " + nginx.Name)

	var operationResult controllerutil.OperationResult
	// Serviceを作成(structの初期化)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: nginx.Namespace,
		},
	}

	service.ObjectMeta.Labels = map[string]string{
		"app":        "nginx",
		"controller": nginx.Name,
	}

	operationResult, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {

		// spec.selectorにlabelsを設定
		if service.Spec.Selector == nil {
			service.Spec.Selector = map[string]string{
				"controller": nginx.Name,
			}
		}

		// spec.portを設定
		if service.Spec.Ports == nil {
			service.Spec.Ports = []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.IntOrString{IntVal: 80},
			}}
		}

		service.Spec.Type = nginx.Spec.ServiceType

		// ★ServiceにOwnerReferenceを設定
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#SetControllerReference
		if err := ctrl.SetControllerReference(nginx, service, r.Scheme); err != nil {
			log.Error(err, "Unable to set OwnerReference from Nginx to Service")
		}

		return nil
	})

	if err != nil {
		log.Error(err, "Unable to ensure service is correct")
		return err
	}

	log.Info("CreateOrUpdate Service for " + nginx.Name + ": " + string(operationResult))

	return nil
}

// OwnerReferenceに設定されたNginxリソースの名前に対応しないDeploymentまたはServiceを削除する
func (r *NginxReconciler) cleanupOwnerResources(ctx context.Context, log logr.Logger, nginx *nginxv1.Nginx) error {
	// log.Info("Finding existing Deployments for Nginx resource")

	/* 以下の条件でDeploymentのListを取得
	   ・Nginxリソースと同じNamespace
	   ・".metadata.controller": <Nginxリソース名>というIndexが付与されている
	*/
	var deploymentList appsv1.DeploymentList
	if err := r.List(ctx, &deploymentList, client.InNamespace(nginx.Namespace), client.MatchingFields(map[string]string{OwnerKey: nginx.Name})); err != nil {
		return err
	}

	for _, deployment := range deploymentList.Items {
		// 取得したDeployment名とNginxで作成されたDeployment名(deploymentName)を比較
		if deployment.Name == nginx.Status.DeploymentName { // trueならDeploymentがNginxに管理されていることになるので削除しない
			// 比較した結果が一致したら何もしない
			continue // 処理をスキップ
		}

		// 比較した結果差分があればDeploymentを削除
		if err := r.Delete(ctx, &deployment); err != nil { // 上でfalseの場合はDeploymentを削除
			log.Error(err, "Faild to delete old Deployment")
			return err
		}
		log.Info("Delete old Deployment resource: " + deployment.Name)

	}

	var serviceList corev1.ServiceList
	if err := r.List(ctx, &serviceList, client.InNamespace(nginx.Namespace), client.MatchingFields(map[string]string{OwnerKey: nginx.Name})); err != nil {
		return err
	}
	for _, service := range serviceList.Items {
		if service.Name == nginx.Status.ServiceName {
			continue
		}

		if err := r.Delete(ctx, &service); err != nil {
			log.Error(err, "Faild to delete old Service")
			return err
		}
		log.Info("Delete old Service resource: " + service.Name)
	}

	return nil
}

//+kubebuilder:rbac:groups=nginx.my.domain,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nginx.my.domain,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nginx.my.domain,resources=nginxes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=services/finalizers,verbs=update

// reconcile.Reconcileインターフェイスを実装
// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	defer fmt.Println("=== End Reconcile " + "for " + req.Name + " ===")
	fmt.Println("=== Start Reconcile " + "for " + req.Name + " ===")

	log := log.FromContext(ctx) // contextに含まれるvalueを付与してログを出力するlogger

	var nginx nginxv1.Nginx
	var deployment appsv1.Deployment
	var service corev1.Service
	var err error

	// ①cacheから変更のあったNginx Objectを取得する
	if err = r.Get(ctx, req.NamespacedName, &nginx); err != nil {
		log.Error(err, "Unable to fetch Nginx from cache.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploymentName := "deploy-" + nginx.Name // Nginxにより管理されるDeploymentの名前
	serviceName := "service-" + nginx.Name   // Nginxにより管理されるServiceの名前

	// ②-1 Nginxが過去に管理していたDeploymentまたはServiceを削除する
	if err = r.cleanupOwnerResources(ctx, log, &nginx); err != nil {
		return ctrl.Result{}, err
	}

	// ③-1 Nginxが管理するDeploymentを作成/更新する
	if err = r.CreateOrUpdateDeployment(ctx, log, &nginx, deploymentName); err != nil {
		return ctrl.Result{}, err
	}

	// ③-2 Nginxが管理するServiceを作成/更新
	if err = r.CreateOrUpdateService(ctx, log, &nginx, serviceName); err != nil {
		return ctrl.Result{}, err
	}

	// ④Nginx ObjectのStatusを更新する
	// controller-runtimeのclientで定義されているObjectKey型でDeploymentのNamespacedNameを設定
	// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/client#ObjectKey
	// ※NamespacedName型が定義できれば良いので以下を直接定義しても多分OK(controller-runtimeがapimachineryをラップしている例)
	// https://pkg.go.dev/k8s.io/apimachinery/pkg/types#NamespacedName

	statusUpdateFlag := false

	deploymentNamespacedName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      deploymentName,
	}

	// NamespacedNameを用いてDeploymentをcacheから取得
	if err = r.Get(ctx, deploymentNamespacedName, &deployment); err != nil {
		log.Error(err, "Unable to fetch Deployment from cache")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Nginx StatusのAvailableReplicasに関する差分比較&更新
	if nginx.Status.AvailableReplicas != deployment.Status.AvailableReplicas {
		nginx.Status.AvailableReplicas = deployment.Status.AvailableReplicas
		statusUpdateFlag = true
	}

	// Nginx StatusのDeploymentNameに関する差分比較&更新
	if nginx.Status.DeploymentName != deployment.Name {
		nginx.Status.DeploymentName = deployment.Name
		statusUpdateFlag = true
	}

	serviceNamespacedName := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      serviceName,
	}

	// NamespacedNameを用いてServiceをcacheから取得
	if err := r.Get(ctx, serviceNamespacedName, &service); err != nil {
		log.Error(err, "Unable to fetch Service from cache")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Nginx StatusのServiceNameに関する差分比較&更新
	if nginx.Status.ServiceName != service.Name {
		nginx.Status.ServiceName = service.Name
		statusUpdateFlag = true
	}

	// Nginx StatusのClusterIPに関する差分比較&更新
	if nginx.Status.ClusterIP != service.Spec.ClusterIP {
		nginx.Status.ClusterIP = service.Spec.ClusterIP
		statusUpdateFlag = true
	}

	// Nginx StatusのExternalIPに関する差分比較&更新
	// service.Status.LoadBalancer.Ingressは配列だけどとりあえず先頭の値をExternalIPとして扱うようにした
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		if nginx.Status.ExternalIP != service.Status.LoadBalancer.Ingress[0].IP {
			nginx.Status.ExternalIP = service.Status.LoadBalancer.Ingress[0].IP
			statusUpdateFlag = true
		}
	} else if (len(service.Status.LoadBalancer.Ingress) == 0) && nginx.Status.ExternalIP != "" {
		nginx.Status.ExternalIP = ""
		statusUpdateFlag = true
	}

	// Nginx Objectの更新(差分ありの場合)
	if statusUpdateFlag {
		log.Info("Update Nginx Status.(nginx.Status.DeploymentName: " + nginx.Status.DeploymentName + ", nginx.Status.AvailableReplicas: " + strconv.Itoa(int(nginx.Status.AvailableReplicas)))
		fmt.Println("  nginx.Status.DeploymentName: " + nginx.Status.DeploymentName)
		fmt.Println("  nginx.Status.AvailableReplicas: " + strconv.Itoa(int(nginx.Status.AvailableReplicas)))
		fmt.Println("  nginx.Status.ServiceName: " + nginx.Status.ServiceName)
		fmt.Println("  nginx.Status.ServiceName: " + nginx.Status.ServiceName)
		fmt.Println("  nginx.Status.ClusterIP: " + nginx.Status.ClusterIP)
		fmt.Println("  nginx.Status.ExternalIP: " + nginx.Status.ExternalIP)
		if err = r.Status().Update(ctx, &nginx); err != nil {
			log.Error(err, "Unable to update Nginx")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// OwnerReferenceの付与状況を確認し、Indexとして付与する値を決める関数
func IndexByOwner(rawObj client.Object) []string {

	var owner *metav1.OwnerReference

	/*
	  アサーションによりclient.Object型として渡されたrawObjの型判定を行い値を取得
	*/
	// rawObjがDeploymentの場合
	if deployment, ok := rawObj.(*appsv1.Deployment); ok {
		// OwnerReferenceへのポインタを取得
		owner = metav1.GetControllerOf(deployment)
	}

	// rawObjがServiceの場合
	if service, ok := rawObj.(*corev1.Service); ok {
		// OwnerReferenceへのポインタを取得
		owner = metav1.GetControllerOf(service)
	}

	// OwnerReferenceが設定されていない場合
	if owner == nil {
		return nil
	}

	if owner.APIVersion != apiGVStr || owner.Kind != "Nginx" {
		return nil
	}

	return []string{owner.Name} // .metadata.controller: owner.NameというIndexを追加(Index Key-ValueのValueを返す)

}

// コントローラー起動時に実行
// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {

	/*
		Controller起動時にcache上のオブジェクトにindexを付与する処理
		https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/client#FieldIndexer.IndexField
		https://book.kubebuilder.io/cronjob-tutorial/controller-implementation.html#setup
		https://zoetrope.github.io/kubebuilder-training/controller-runtime/manager.html

		IndexFieldはクラスタ上に存在する全てのDeployment(第2引数で指定したリソース)に対してindex付与の処理を行う
		  第2引数: indexの付与対象リソース(appsv1.Deployment)
		  		  第2引数はIndexFieldの仕様上、client.Object (interface)をとる
				  ここに&appsv1.Deployment{} (struct)を指定することで
				    interface = &struct
				  という形になり、実装を行っているようなイメージ
		  第3引数: indexのKey(deploymentOwnerKey=.metadata.controller)を指定(任意の文字列)
		  第4引数: indexのKeyに対するValueを決める関数(https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/client#IndexerFunc)
		          funcでの戻り値が弟3引数のdeploymentOwnerKeyというindex Keyのvalueとしてセットされる
				  クラスタ上に存在する&appsv1.Deploymentに対してclient.Objectを実装したもの(第2引数)を順次func(rawObj client.Object)に入れて処理していく感じ
	*/
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, OwnerKey, IndexByOwner); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, OwnerKey, IndexByOwner); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&nginxv1.Nginx{}).
		Owns(&appsv1.Deployment{}). // Controllerに作成されるリソースを指定
		Owns(&corev1.Service{}).
		Complete(r)
}
