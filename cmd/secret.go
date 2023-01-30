package cmd

import (
	"context"
	"fmt"
	"strings"

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
	fmt.Println("Secret", secret.Name, "has been deleted")
}

func createOrUpdate(obj interface{}) {
	secret := obj.(*corev1.Secret)
	srcSecretNamespace := strings.Split(srcSecret, "/")[0]
	srcSecretName := strings.Split(srcSecret, "/")[1]
	destSecretNamespace := strings.Split(destSecret, "/")[0]
	destSecretName := strings.Split(destSecret, "/")[1]

	isExists := true
	if secret.Namespace == srcSecretNamespace && secret.Name == srcSecretName {
		s, getErr := client.CoreV1().Secrets(destSecretNamespace).Get(context.TODO(), destSecretName, metav1.GetOptions{})
		if errors.IsNotFound(getErr) {
			// if not exists, create
			isExists = false
			klog.Info("cannot found Secret: ", destSecretName)
			_, err := client.CoreV1().Secrets(destSecretNamespace).Create(context.TODO(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      destSecretName,
					Namespace: destSecretNamespace,
				},
				Data: secret.Data,
			}, metav1.CreateOptions{})
			if err != nil {
				panic(fmt.Errorf("failed to create Secret: %v", getErr))
			}
			klog.Info("successfully create Secret: ", destSecretName)
		} else if getErr != nil {
			panic(fmt.Errorf("failed to get Secret: %v", getErr))
		}
		// if exists, update
		if isExists {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				s.Data = secret.Data
				_, err := client.CoreV1().Secrets(destSecretNamespace).Update(context.TODO(), s, metav1.UpdateOptions{})
				return err
			})
			if retryErr != nil {
				panic(fmt.Errorf("failed to update Secret: %v", retryErr))
			}
			klog.Info("successfully update Secret: ", destSecretName)
		}
	}
}
