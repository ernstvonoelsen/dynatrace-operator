package daemonset

import (
	dynatracev1beta1 "github.com/Dynatrace/dynatrace-operator/src/api/v1beta1"
	dtcsi "github.com/Dynatrace/dynatrace-operator/src/controllers/csi"
	csivolumes "github.com/Dynatrace/dynatrace-operator/src/controllers/csi/driver/volumes"
	hostvolumes "github.com/Dynatrace/dynatrace-operator/src/controllers/csi/driver/volumes/host"
	corev1 "k8s.io/api/core/v1"
)

func prepareVolumeMounts(instance *dynatracev1beta1.DynaKube) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	if instance.NeedsReadOnlyOneAgents() {
		volumeMounts = append(volumeMounts, getReadOnlyRootMount())
		volumeMounts = append(volumeMounts, getCSIStorageMount())

	} else {
		volumeMounts = append(volumeMounts, getRootMount())
	}

	if instance.Spec.ClusterCAs != "" {
		volumeMounts = append(volumeMounts, getClusterCaCertificateMount())
	}

	if instance.HasActiveGateCustomCa() {
		volumeMounts = append(volumeMounts, getActiveGateCaMount())
	}
	return volumeMounts
}

func getClusterCaCertificateMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      clusterCaVolumeName,
		MountPath: clusterCaVolumeMount,
	}
}

func getActiveGateCaMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      agCaVolumeName,
		MountPath: agCaVolumeMount,
	}
}

func getRootMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      hostRootVolumeName,
		MountPath: hostRootVolumeMount,
	}
}

func getReadOnlyRootMount() corev1.VolumeMount {
	rootMount := getRootMount()
	rootMount.ReadOnly = true
	return rootMount
}

func getCSIStorageMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      csiStorageVolumeName,
		MountPath: csiStorageVolumeMount,
	}
}

func prepareVolumes(instance *dynatracev1beta1.DynaKube) []corev1.Volume {
	volumes := []corev1.Volume{getRootVolume()}

	if instance.NeedsReadOnlyOneAgents() {
		volumes = append(volumes, getCSIStorageVolume(instance))
	}

	if instance.Spec.ClusterCAs != "" {
		volumes = append(volumes, getCertificateVolume(instance))
	}

	if instance.HasActiveGateCustomCa() {
		volumes = append(volumes, getTLSVolume(instance))
	}

	return volumes
}

func getCertificateVolume(instance *dynatracev1beta1.DynaKube) corev1.Volume {
	return corev1.Volume{
		Name: clusterCaVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: instance.Spec.ClusterCAs,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "certs",
						Path: "certs.pem",
					},
				},
			},
		},
	}
}

func getCSIStorageVolume(instance *dynatracev1beta1.DynaKube) corev1.Volume {
	return corev1.Volume{
		Name: csiStorageVolumeName,
		VolumeSource: corev1.VolumeSource{
			CSI: &corev1.CSIVolumeSource{
				Driver: dtcsi.DriverName,
				VolumeAttributes: map[string]string{
					csivolumes.CSIVolumeAttributeModeField:     hostvolumes.Mode,
					csivolumes.CSIVolumeAttributeDynakubeField: instance.Name,
				},
			},
		},
	}
}

func getTLSVolume(instance *dynatracev1beta1.DynaKube) corev1.Volume {
	return corev1.Volume{
		Name: agCaVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: instance.Spec.ActiveGate.TlsSecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  "server.crt",
						Path: "custom.pem",
					},
				},
			},
		},
	}
}

func getRootVolume() corev1.Volume {
	return corev1.Volume{
		Name: hostRootVolumeName,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/",
			},
		},
	}
}
