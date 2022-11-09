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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var nginxlog = logf.Log.WithName("nginx-resource")

func (r *Nginx) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-nginx-my-domain-v1-nginx,mutating=true,failurePolicy=fail,sideEffects=None,groups=nginx.my.domain,resources=nginxes,verbs=create;update,versions=v1,name=mnginx.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Nginx{}

// Mutation
// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Nginx) Default() {
	nginxlog.Info("[Mutation] Start Mutation", "name", r.Name)
	nginxlog.Info("[Mutation] Add Annotations: nginx: "+r.Name, "name", r.Name)

	// Nginxリソース作成時にAnnotationsを付与する
	annotations := make(map[string]string)
	annotations["nginx"] = r.Name
	r.ObjectMeta.Annotations = annotations

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-nginx-my-domain-v1-nginx,mutating=false,failurePolicy=fail,sideEffects=None,groups=nginx.my.domain,resources=nginxes,verbs=create;update,versions=v1,name=vnginx.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Nginx{}

// Nginxリソース名の文字数を確認するメソッド
func (r *Nginx) validateNginxName() error {
	nginxlog.Info("[Validation] Check Nginx charactors", "name", r.Name)

	var errs field.ErrorList

	// 20文字以上ならエラー
	if len(r.ObjectMeta.Name) > 20 {
		errs = append(errs, field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 20 characters."))
	}

	// Validation Webhookに失敗したらerrors.StatusError型のエラーを返す
	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: "nginx", Kind: "Nginx"}, r.Name, errs)
		nginxlog.Error(err, "validation error", "name", r.Name)
		return err
	}

	return nil
}

// Validation
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Nginx) ValidateCreate() error {
	nginxlog.Info("[Validation] Validate Create", "name", r.Name)

	return r.validateNginxName()

}

// Validation
// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Nginx) ValidateUpdate(old runtime.Object) error {
	nginxlog.Info("[Validation] Validate Update", "name", r.Name)

	return r.validateNginxName()
}

// Validation
// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Nginx) ValidateDelete() error {
	nginxlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
