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
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sdboperatorv1alpha1 "github.com/razzie/k8s-db-operator/pkg/api/v1alpha1"
	"github.com/razzie/k8s-db-operator/pkg/postgres"
	"github.com/razzie/k8s-db-operator/pkg/redis"
)

var (
	ErrUnknownDatabaseType = fmt.Errorf("unknown database type")
)

// DatabaseClaimReconciler reconciles a DatabaseClaim object
type DatabaseClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s-db-operator.razzie.github.io,resources=databaseclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s-db-operator.razzie.github.io,resources=databaseclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s-db-operator.razzie.github.io,resources=databaseclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseClaim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *DatabaseClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	dbclaim := &k8sdboperatorv1alpha1.DatabaseClaim{}
	if err := r.Get(ctx, req.NamespacedName, dbclaim); err != nil {
		log.Error(err, "unable to fetch object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !dbclaim.Status.Ready {
		connStr, err := createNewConnectionString(ctx, dbclaim.Spec.DatabaseType)
		if err != nil {
			if errors.Is(err, ErrUnknownDatabaseType) {
				log.Error(err, "", "spec.databaseType", string(dbclaim.Spec.DatabaseType))
				return ctrl.Result{}, err
			}
			log.Error(err, "failed to create new connection string")
			return ctrl.Result{Requeue: true}, err
		}

		// create secret
		secret := &corev1.Secret{}
		secret.ObjectMeta.Name = dbclaim.Spec.SecretName
		secret.ObjectMeta.Namespace = req.Namespace
		secret.Data = map[string][]byte{
			"connectionString": []byte(connStr),
		}
		controllerutil.SetOwnerReference(dbclaim, secret, r.Scheme)
		err = r.Client.Create(ctx, secret)
		if err != nil {
			log.Error(err, "failed to create secret")
			return ctrl.Result{}, err
		}
		log.Info("created secret", "secret", secret)

		// update cr state to ready
		dbclaim.Status.Ready = true
		r.Client.Status().Update(ctx, dbclaim)
		log.Info("updated ready status to true", "dbclaim", req.NamespacedName)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sdboperatorv1alpha1.DatabaseClaim{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func createNewConnectionString(ctx context.Context, dbType k8sdboperatorv1alpha1.DatabaseType) (string, error) {
	switch dbType {
	case "PostgreSQL":
		return postgres.CreateNewConnectionString(ctx)
	case "Redis":
		return redis.CreateNewConnectionString(ctx)
	default:
		return "", ErrUnknownDatabaseType
	}
}
