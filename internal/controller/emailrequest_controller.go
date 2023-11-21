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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hiringv1alpha1 "github.com/cannonpalms/email-controller-template/api/v1alpha1"
	"github.com/cannonpalms/email-controller-template/pkg/fakeemail"
)

// EmailRequestReconciler reconciles a EmailRequest object
type EmailRequestReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	EmailService *fakeemail.EmailService
}

/*
	// true when sent successfully (or reached a nonretryable error), otherwise false
	SendSuccess bool `json:"sendSuccess"`
	// the last error received on a send attempt or why it won't be retried, empty if it has never been sent
	LastError string `json:"lastError"`
	// time the status was last updated, for informational purposes only.
	UpdatedAt string `json:"updatedAt"`
*/

const EmailRequestStatusType = "EmailRequestSentSuccess"

//+kubebuilder:rbac:groups=hiring.influxdata.io,resources=emailrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hiring.influxdata.io,resources=emailrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hiring.influxdata.io,resources=emailrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EmailRequest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *EmailRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var emailRequest hiringv1alpha1.EmailRequest
	fmt.Println("Reconciling EmailRequest")
	err := r.Get(ctx, req.NamespacedName, &emailRequest)
	if err != nil {
		fmt.Println("Error getting emailRequest", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get the status condition - only one condition
	var condition *metav1.Condition
	if len(emailRequest.Status.Conditions) == 0 {
		condition = &metav1.Condition{
			Type:               EmailRequestStatusType,
			Status:             metav1.ConditionFalse,
			Reason:             "EmailRequestStatusInitialized",
			Message:            "EmailRequest status initialized",
			LastTransitionTime: metav1.Now(),
		}

		// initialize the status
		emailRequest.Status.Conditions = []metav1.Condition{
			*condition,
		}
		err = r.Status().Update(ctx, &emailRequest)
		if err != nil {
			fmt.Println("Error patching emailRequest status", err)
			return ctrl.Result{}, err
		}
	} else {
		condition = &emailRequest.Status.Conditions[0]
	}

	if condition.Status == metav1.ConditionTrue {
		fmt.Println("Email already sent successfully")
		return ctrl.Result{}, nil
	}

	err = r.EmailService.Send(emailRequest.Spec.Address, "Hello", fmt.Sprintf("Hello, %s!", emailRequest.Spec.Name))
	var retry bool
	var statusMsg string
	var reason string
	switch errType := err.(type) {
	case *fakeemail.ErrInvalidEmailAddress:
		retry = false
		statusMsg = fmt.Sprintf("Invalid email address: %s", errType.Email)
		reason = "InvalidEmailAddress"
	case *fakeemail.ErrEmailBounced:
		retry = false
		statusMsg = fmt.Sprintf("Email bounced: %s", errType.Email)
		reason = "EmailBounced"
	case *fakeemail.ErrEmailBlocked:
		retry = emailRequest.Spec.RetryBlockedPolicy
		statusMsg = fmt.Sprintf("Email blocked: %s", errType.Email)
		reason = "EmailBlocked"
	default:
		if err != nil {
			// unknown other error
			fmt.Println("Unknown error sending email", err)
			return ctrl.Result{}, err
		}
		reason = "EmailSent"
	}

	fmt.Println(statusMsg)
	if retry {
		condition = &metav1.Condition{
			Type:               EmailRequestStatusType,
			Status:             metav1.ConditionFalse,
			Reason:             "RetryingSendFailure",
			Message:            statusMsg,
			LastTransitionTime: metav1.Now(),
		}
	} else {
		condition = &metav1.Condition{
			Type:               EmailRequestStatusType,
			Status:             metav1.ConditionTrue,
			Reason:             reason,
			Message:            statusMsg,
			LastTransitionTime: metav1.Now(),
		}
	}

	emailRequest.Status.Conditions = []metav1.Condition{
		*condition,
	}
	err = r.Status().Update(ctx, &emailRequest)
	if err != nil {
		fmt.Println("Error patching emailRequest", err)
		return ctrl.Result{}, err
	}

	fmt.Println("Reconcile complete: retry = ", retry)
	if retry {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: time.Minute,
		}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmailRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hiringv1alpha1.EmailRequest{}).
		Complete(r)
}
