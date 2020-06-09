package kafka

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdateClientService(ctx context.Context, c client.Client, newService *corev1.Service) (controllerutil.OperationResult, error) {
	key, err := client.ObjectKeyFromObject(newService)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	var oldService = &corev1.Service{}
	if err := c.Get(ctx, key, oldService); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, newService); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	if newService.Spec.Ports[0].Port != oldService.Spec.Ports[0].Port {
		oldService.Spec.Ports[0].Port = newService.Spec.Ports[0].Port
		if err := c.Update(ctx, oldService); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultUpdated, nil
	}
	return controllerutil.OperationResultNone, nil
}

func CreateOrUpdateHeadlessService(ctx context.Context, c client.Client, newService *corev1.Service) (controllerutil.OperationResult, error) {
	key, err := client.ObjectKeyFromObject(newService)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	var oldService = &corev1.Service{}
	if err := c.Get(ctx, key, oldService); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, newService); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	if newService.Spec.Ports[0].Port != oldService.Spec.Ports[0].Port {
		oldService.Spec.Ports[0].Port = newService.Spec.Ports[0].Port
		oldService.Spec.Ports[0].TargetPort = intstr.IntOrString{
			IntVal: newService.Spec.Ports[0].Port,
		}
		if err := c.Update(ctx, oldService); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultUpdated, nil
	}
	return controllerutil.OperationResultNone, nil
}

func CreateOrUpdatePodDisruptionBudget(ctx context.Context, c client.Client, newPdb *policy.PodDisruptionBudget) (controllerutil.OperationResult, error) {
	key, err := client.ObjectKeyFromObject(newPdb)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	var oldPdb = &policy.PodDisruptionBudget{}
	if err := c.Get(ctx, key, oldPdb); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, newPdb); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	if newPdb.Spec.MinAvailable.IntVal != oldPdb.Spec.MinAvailable.IntVal {
		oldPdb.Spec.MinAvailable.IntVal = newPdb.Spec.MinAvailable.IntVal
		if err := c.Update(ctx, oldPdb); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultUpdated, nil
	}
	return controllerutil.OperationResultNone, nil
}

func CreateOrUpdateStatefulSet(ctx context.Context, c client.Client, newStatefulSet *appsv1.StatefulSet) (controllerutil.OperationResult, error) {
	key, err := client.ObjectKeyFromObject(newStatefulSet)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	var oldStatefulSet = &appsv1.StatefulSet{}
	if err := c.Get(ctx, key, oldStatefulSet); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, newStatefulSet); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	newVolume := newStatefulSet.Spec.VolumeClaimTemplates
	oldVolume := oldStatefulSet.Spec.VolumeClaimTemplates
	if newVolume == nil && oldVolume == nil {
		return updatePublicInfo(ctx, c, newStatefulSet, oldStatefulSet)
	} else if newVolume != nil && oldVolume == nil {
		oldStatefulSet.Spec.VolumeClaimTemplates = newStatefulSet.Spec.VolumeClaimTemplates
		oldStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = newStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
		if err := c.Update(ctx, oldStatefulSet); err != nil {
			return controllerutil.OperationResultNone, err
		}
		result, err := updatePublicInfo(ctx, c, newStatefulSet, oldStatefulSet)
		if err == nil && result == controllerutil.OperationResultNone {
			return controllerutil.OperationResultUpdated, nil
		}
	} else if newVolume == nil && oldVolume != nil {
		oldStatefulSet.Spec.VolumeClaimTemplates = newStatefulSet.Spec.VolumeClaimTemplates
		oldStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = newStatefulSet.Spec.Template.Spec.Containers[0].VolumeMounts
		if err := c.Update(ctx, oldStatefulSet); err != nil {
			return controllerutil.OperationResultNone, err
		}
		result, err := updatePublicInfo(ctx, c, newStatefulSet, oldStatefulSet)
		if err == nil && result == controllerutil.OperationResultNone {
			return controllerutil.OperationResultUpdated, nil
		}
	} else if newVolume != nil && oldVolume != nil {
		if newVolume[0].Spec.StorageClassName != oldVolume[0].Spec.StorageClassName ||
			newVolume[0].Spec.Resources.Requests[corev1.ResourceStorage] != oldVolume[0].Spec.Resources.Requests[corev1.ResourceStorage] {
			oldVolume[0].Spec.StorageClassName = newVolume[0].Spec.StorageClassName
			oldVolume[0].Spec.Resources.Requests[corev1.ResourceStorage] = newVolume[0].Spec.Resources.Requests[corev1.ResourceStorage]
			if err := c.Update(ctx, oldStatefulSet); err != nil {
				return controllerutil.OperationResultNone, err
			}
			result, err := updatePublicInfo(ctx, c, newStatefulSet, oldStatefulSet)
			if err == nil && result == controllerutil.OperationResultNone {
				return controllerutil.OperationResultUpdated, nil
			}
		} else {
			return updatePublicInfo(ctx, c, newStatefulSet, oldStatefulSet)
		}
	}

	return controllerutil.OperationResultNone, nil
}

func updatePublicInfo(ctx context.Context, c client.Client, newStatefulSet *appsv1.StatefulSet, oldStatefulSet *appsv1.StatefulSet) (controllerutil.OperationResult, error) {
	newContainer := newStatefulSet.Spec.Template.Spec.Containers[0]
	oldContainer := oldStatefulSet.Spec.Template.Spec.Containers[0]
	if *newStatefulSet.Spec.Replicas != *oldStatefulSet.Spec.Replicas ||
		newContainer.Ports[0].ContainerPort != oldContainer.Ports[0].ContainerPort ||
		newContainer.Env[0].Value != oldContainer.Env[0].Value ||
		newContainer.Env[1].Value != oldContainer.Env[1].Value {

		oldStatefulSet.Spec.Replicas = newStatefulSet.Spec.Replicas
		oldContainer.Ports[0].ContainerPort = newContainer.Ports[0].ContainerPort
		oldContainer.Env[0].Value = newContainer.Env[0].Value
		oldContainer.Env[1].Value = newContainer.Env[1].Value

		if err := c.Update(ctx, oldStatefulSet); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultUpdated, nil
	}
	return controllerutil.OperationResultNone, nil
}
