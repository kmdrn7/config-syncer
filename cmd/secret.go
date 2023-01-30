package cmd

import (
	"config-syncer/pkg/config"
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

func secretOnAdd(obj interface{}) {
	createOrUpdate(obj)
}

func secretOnUpdate(obj interface{}, new interface{}) {
	createOrUpdate(new)
}

func secretOnDelete(obj interface{}) {
	secret := obj.(*corev1.Secret)
	klog.Infof("secret %s/%s has been deleted", secret.Namespace, secret.Name)
}

func createOrUpdate(obj interface{}) {
	secret := obj.(*corev1.Secret)
	appConfig := config.GetConfig()

	secretFound := false
	for _, sec := range appConfig.Secrets {
		if secret.Namespace == sec.Namespace && secret.Name == sec.Name {
			secretFound = true
			klog.Infof("processing secret %s/%s ", secret.Namespace, secret.Name)
			for _, dest := range sec.Destinations {
				isExists := true
				request, getErr := client.CoreV1().Secrets(dest.Namespace).Get(context.TODO(), dest.Name, metav1.GetOptions{})
				if errors.IsNotFound(getErr) {
					// if not exists, create
					isExists = false
					klog.Infof("cannot find secret %s/%s ", dest.Namespace, dest.Name)
					_, err := client.CoreV1().Secrets(dest.Namespace).Create(context.TODO(), &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      dest.Name,
							Namespace: dest.Namespace,
						},
						Data: secret.Data,
					}, metav1.CreateOptions{})
					if err != nil {
						klog.Errorf("failed to create secret: %v", err)
					}
					klog.Info("successfully create Secret: ", dest.Name)
				} else if getErr != nil {
					klog.Errorf("failed to get secret: %v", getErr)
				}
				// if exists, update
				if isExists {
					retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
						request.Data = secret.Data
						_, err := client.CoreV1().Secrets(dest.Namespace).Update(context.TODO(), request, metav1.UpdateOptions{})
						return err
					})
					if retryErr != nil {
						klog.Errorf("failed to update secret: %v", retryErr)
					}
					klog.Info("successfully update Secret: ", dest.Name)
				}
			}
		}
	}

	if !secretFound {
		if debug {
			klog.Infof("skip processing secret %s/%s", secret.Namespace, secret.Name)
		}
	}
}
