readonly CRIO_HOOKS_DIR="/usr/share/containers/oci/hooks.d"
readonly CRIO_HOOK_FILENAME="oci-xdxct-hook.json"

# shellcheck disable=SC2015
[ -t 2 ] && readonly LOG_TTY=1 || readonly LOG_NO_TTY=1

if [ "${LOG_TTY-0}" -eq 1 ] && [ "$(tput colors)" -ge 15 ]; then
	readonly FMT_BOLD=$(tput bold)
	readonly FMT_RED=$(tput setaf 1)
	readonly FMT_YELLOW=$(tput setaf 3)
	readonly FMT_BLUE=$(tput setaf 12)
	readonly FMT_CLEAR=$(tput sgr0)
fi

log() {
	local -r level="$1"; shift
	local -r message="$*"

	local fmt_on="${FMT_CLEAR-}"
	local -r fmt_off="${FMT_CLEAR-}"

	case "${level}" in
		INFO)  fmt_on="${FMT_BLUE-}" ;;
		WARN)  fmt_on="${FMT_YELLOW-}" ;;
		ERROR) fmt_on="${FMT_RED-}" ;;
	esac
	printf "%s[%s]%s %b\n" "${fmt_on}" "${level}" "${fmt_off}" "${message}" >&2
}

with_retry() {
	local max_attempts="$1"
	local delay="$2"
	local count=0
	local rc
	shift 2

	while true; do
		set +e
		"$@"; rc="$?"
		set -e

		count="$((count+1))"

		if [[ "${rc}" -eq 0 ]]; then
			return 0
		fi

		if [[ "${max_attempts}" -le 0 ]] || [[ "${count}" -lt "${max_attempts}" ]]; then
			sleep "${delay}"
		else
			break
		fi
	done

	return 1
}

testing::setup() {
	cp -Rp ${basedir}/shared ${shared_dir}
	mkdir -p "${shared_dir}/etc/containerd"
	mkdir -p "${shared_dir}/etc/docker"
	mkdir -p "${shared_dir}/run/docker/containerd"
	mkdir -p "${shared_dir}/run/xdxct"
	mkdir -p "${shared_dir}/usr/local/xdxct"
	mkdir -p "${shared_dir}${CRIO_HOOKS_DIR}"
}

testing::cleanup() {
	if [[ "${CLEANUP}" == "false" ]]; then
		echo "Skipping cleanup: CLEANUP=${CLEANUP}"
		return 0
	fi
	if [[ -e "${shared_dir}" ]]; then
		docker run --rm \
			-v "${shared_dir}:/work" \
			alpine sh -c 'rm -rf /work/*'
		rmdir "${shared_dir}"
	fi

	if [[ "${test_cases:-""}" == "" ]]; then
		echo "No test cases defined. Skipping test case cleanup"
		return 0
	fi

	for tc in ${test_cases}; do
		testing::${tc}::cleanup
	done
}

testing::docker_run::toolkit::shell() {
	docker run --rm --privileged \
		--entrypoint sh \
		-v "${shared_dir}/etc/containerd:/etc/containerd" \
		-v "${shared_dir}/etc/docker:/etc/docker" \
		-v "${shared_dir}/run/docker/containerd:/run/docker/containerd" \
		-v "${shared_dir}/run/xdxct:/run/xdxct" \
		-v "${shared_dir}/usr/local/xdxct:/usr/local/xdxct" \
		-v "${shared_dir}${CRIO_HOOKS_DIR}:${CRIO_HOOKS_DIR}" \
		"${toolkit_container_image}" "-c" "$*"
}


