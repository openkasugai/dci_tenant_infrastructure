#!/bin/sh -u
die() { echo "[ERROR] $*" >&2; exit 1; }

root_dir=$(cd "$(dirname "$0")/.." && pwd)
dist_dir="dist" # relative path from root_dir
kubebuilder_bin="${KUBEBUILDER_BIN:-kubebuilder}"
build_installer_target="${BUILD_INSTALLER_TARGET:-build-installer-alert-receiver}"
install_manifest="${dist_dir}/install.yaml"
tmp_dir=$(mktemp -d)
install_scale_controller_manifest="${dist_dir}/install-scale-controller.yaml"
install_alert_receiver_manifest="${dist_dir}/install-alert-receiver.yaml"

cleanup() {
	rm -rf "${tmp_dir}"
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || die "$1 command not found"
}

build_install_manifest() {
	make "${build_installer_target}" || die "make ${build_installer_target} failed"
	[ -f "${install_manifest}" ] || die "install manifest not found: ${install_manifest}"
}

split_install_manifests_for_alert_receiver() {
	# Keep install.yaml read-only and create two derived manifests:
	# - alert-receiver related resources
	# - everything else
	rm -f "${install_scale_controller_manifest}" "${install_alert_receiver_manifest}"
	yq eval '
		select(((.metadata.name // "") | test("alert-receiver")))
	' "${install_manifest}" > "${install_alert_receiver_manifest}" || die "failed to create install-alert-receiver.yaml"
	yq eval '
		select(((.metadata.name // "") | test("alert-receiver")) | not)
	' "${install_manifest}" > "${install_scale_controller_manifest}" || die "failed to create install-controller-manager.yaml"
}

generate_chart_from_kubebuilder() {
	rm -rf "${dist_dir}/chart"
	# We intentionally skip pre-checking kubebuilder version here and rely on
	# `kubebuilder edit` to fail when the installed version is unsupported.
	# helm/v2-alpha expects kubebuilder v4.10.0+ (recommended: v4.10.1+).
	[ -f "${install_scale_controller_manifest}" ] || die "manifests file not found: ${install_scale_controller_manifest}"
	"${kubebuilder_bin}" edit --plugins=helm/v2-alpha --manifests="${install_scale_controller_manifest}" || die "kubebuilder edit failed"
}

generate_alert_receiver_chart_templates() {
	if [ ! -s "${install_alert_receiver_manifest}" ]; then
		return 0
	fi

	alert_receiver_template_dir="${dist_dir}/chart/templates/alert-receiver"
	alert_receiver_tmp_rendered="${tmp_dir}/alert-receiver-rendered.yaml"
	rm -rf "${alert_receiver_template_dir}"
	mkdir -p "${alert_receiver_template_dir}"

	yq eval '
		select(.kind != null)
		| .metadata.labels."app.kubernetes.io/managed-by" = "{{ .Release.Service }}"
		| .metadata.labels."app.kubernetes.io/version" = "{{ .Chart.AppVersion }}"
		| .metadata.labels."app.kubernetes.io/instance" = "{{ .Release.Name }}"
		| .metadata.labels."helm.sh/chart" = "{{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}"
		| with(select(.metadata.namespace != null); .metadata.namespace = "{{ .Release.Namespace }}")
		| with(select(.kind == "ServiceAccount"); .metadata.name = "{{ .Values.alertReceiver.serviceAccountName }}")
		| with(select(.kind == "Deployment");
			.metadata.name = "{{ include \"scale-controller.resourceName\" (dict \"suffix\" \"alert-receiver\" \"context\" $) }}" |
			.spec.replicas = "{{ .Values.alertReceiver.replicas }}" |
			.spec.selector.matchLabels."app.kubernetes.io/instance" = "{{ .Release.Name }}" |
			.spec.template.metadata.labels."app.kubernetes.io/managed-by" = "{{ .Release.Service }}" |
			.spec.template.metadata.labels."app.kubernetes.io/version" = "{{ .Chart.AppVersion }}" |
			.spec.template.metadata.labels."app.kubernetes.io/instance" = "{{ .Release.Name }}" |
			.spec.template.metadata.labels."helm.sh/chart" = "{{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}" |
			.spec.template.spec.serviceAccountName = "{{ .Values.alertReceiver.serviceAccountName }}" |
			.spec.template.spec.containers[0].env[0].name = "CARD_ALERT_NAMESPACE" |
			.spec.template.spec.containers[0].env[0].value = "{{ .Values.alertReceiver.container.env.CARD_ALERT_NAMESPACE }}" |
			.spec.template.spec.containers[0].image = "{{ .Values.alertReceiver.container.image.repository }}:{{ .Values.alertReceiver.container.image.tag }}"
		  )
		| with(select(.kind == "ClusterRoleBinding");
			.subjects[0].name = "{{ .Values.alertReceiver.serviceAccountName }}" |
			.subjects[0].namespace = "{{ .Release.Namespace }}"
		  )
	' "${install_alert_receiver_manifest}" > "${alert_receiver_tmp_rendered}" || die "failed to render alert-receiver resources with yq"

	# Keep replicas as a numeric Helm expression instead of a quoted string.
	sed -i "s|^  replicas: '{{ .Values.alertReceiver.replicas }}'$|  replicas: {{ .Values.alertReceiver.replicas }}|" "${alert_receiver_tmp_rendered}" \
		|| die "failed to normalize alert-receiver replicas"
	grep -q "^  replicas: {{ .Values.alertReceiver.replicas }}$" "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify alert-receiver replicas normalization"

	# Normalize label quoting style:
	# - quoted: app.kubernetes.io/version, helm.sh/chart
	# - unquoted: app.kubernetes.io/managed-by, app.kubernetes.io/instance
	sed -i \
		-e "s@app.kubernetes.io/managed-by: '{{ .Release.Service }}'@app.kubernetes.io/managed-by: {{ .Release.Service }}@g" \
		-e "s@app.kubernetes.io/instance: '{{ .Release.Name }}'@app.kubernetes.io/instance: {{ .Release.Name }}@g" \
		-e "s@app.kubernetes.io/version: '{{ .Chart.AppVersion }}'@app.kubernetes.io/version: \"{{ .Chart.AppVersion }}\"@g" \
		-e "s@helm.sh/chart: '{{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}'@helm.sh/chart: \"{{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}\"@g" \
		"${alert_receiver_tmp_rendered}" || die "failed to normalize alert-receiver label quotes"
	grep -q "app.kubernetes.io/managed-by: {{ .Release.Service }}" "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify managed-by label normalization"
	grep -q "app.kubernetes.io/instance: {{ .Release.Name }}" "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify instance label normalization"
	grep -q 'app.kubernetes.io/version: "{{ .Chart.AppVersion }}"' "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify version label normalization"
	grep -q 'helm.sh/chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"' "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify chart label normalization"

	# Keep alert-receiver metadata name/namespace unquoted.
	sed -i \
		-e "s#^  name: '{{ .Values.alertReceiver.serviceAccountName }}'\$#  name: {{ .Values.alertReceiver.serviceAccountName }}#" \
		-e "s#^  name: '{{ include \"scale-controller.resourceName\" (dict \"suffix\" \"alert-receiver\" \"context\" \\$) }}'\$#  name: {{ include \"scale-controller.resourceName\" (dict \"suffix\" \"alert-receiver\" \"context\" \\$) }}#" \
		-e "s#^  namespace: '{{ .Release.Namespace }}'\$#  namespace: {{ .Release.Namespace }}#" \
		"${alert_receiver_tmp_rendered}" || die "failed to normalize alert-receiver metadata name/namespace quotes"
	grep -q "^  name: {{ .Values.alertReceiver.serviceAccountName }}$" "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify serviceaccount name normalization"
	grep -q '^  name: {{ include "scale-controller.resourceName" (dict "suffix" "alert-receiver" "context" \$) }}$' "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify deployment name normalization"
	grep -q "^  namespace: {{ .Release.Namespace }}$" "${alert_receiver_tmp_rendered}" \
		|| die "failed to verify namespace normalization"

	{
		echo '{{- if .Values.alertReceiver.enabled }}'
		cat "${alert_receiver_tmp_rendered}"
		echo '{{- end -}}'
	} > "${alert_receiver_template_dir}/resources.yaml" || die "failed to write alert-receiver/resources.yaml"
}

copy_chart_patches() {
	patch_src_dir="helm/patches/chart"
	patch_list_file="${tmp_dir}/chart-patches.list"

	[ -d "${patch_src_dir}" ] || return 0
	find "${patch_src_dir}" \( -name "*.yaml" -o -name "*.tpl" \) | LC_ALL=C sort > "${patch_list_file}" \
		|| die "failed to enumerate chart patch files"

	while IFS= read -r yaml_file; do
		[ -n "${yaml_file}" ] || continue
		# Preserve relative path from helm/patches/chart
		rel_path="${yaml_file#${patch_src_dir}/}"
		target_file="${dist_dir}/chart/${rel_path}"
		patch_out="${patch_src_dir}/${rel_path}.patch"
		mkdir -p "$(dirname "${patch_out}")"

		# Create unified diff (handle missing target as new file)
		if [ -f "${target_file}" ]; then
			diff -u "${target_file}" "${yaml_file}" > "${patch_out}"; rc=$?
		else
			diff -u /dev/null "${yaml_file}" > "${patch_out}"; rc=$?
		fi
		if [ $rc -eq 0 ]; then
			# No differences; remove empty patch file
			rm -f "${patch_out}"
		elif [ $rc -eq 1 ]; then
			# Differences exist; keep patch
			echo "Created patch for ${rel_path} at ${patch_out}"
		else
			die "diff failed creating patch for ${rel_path} (rc=${rc})"
		fi

		# Copy YAML into dist preserving directory structure
		mkdir -p "$(dirname "${target_file}")"
		cp -f "${yaml_file}" "${target_file}" || die "Failed to copy ${yaml_file} to ${target_file}"
	done < "${patch_list_file}"
}

package_chart() {
	helm package "${dist_dir}/chart" --destination "${dist_dir}" || die "helm package failed"
}

main() {
	trap cleanup EXIT
	cd "${root_dir}" || die "Failed to enter root_dir: ${root_dir}"

	for required_cmd in "${kubebuilder_bin}" yq make diff helm
	do
		require_command "${required_cmd}"
	done

	build_install_manifest
	split_install_manifests_for_alert_receiver

	generate_chart_from_kubebuilder
	generate_alert_receiver_chart_templates

	copy_chart_patches
	package_chart
}

main "$@"
