readonly containerd_dind_ctr="container-config-containerd-dind-ctr-name"
readonly containerd_test_ctr="container-config-containerd-test-ctr-name"
readonly containerd_dind_socket="/run/xdxct/docker.sock"
readonly containerd_dind_containerd_dir="/run/docker/containerd"

testing::containerd::dind::setup() {
	# Docker creates /etc/docker when starting
	# by default there isn't any config in this directory (even after the daemon starts)
	docker run -d --rm --privileged \
		-v "${shared_dir}/etc/docker:/etc/docker" \
		-v "${shared_dir}/run/xdxct:/run/xdxct" \
		-v "${shared_dir}/usr/local/xdxct:/usr/local/xdxct" \
		-v "${shared_dir}/run/docker/containerd:/run/docker/containerd" \
		--name "${containerd_dind_ctr}" \
		docker:stable-dind -H unix://${containerd_dind_socket}
}

testing::containerd::dind::exec() {
	docker exec "${containerd_dind_ctr}" sh -c "$*"
}

testing::containerd::toolkit::run() {
	local version=${1}

	# We run ctr image list to ensure that containerd has successfully started in the docker-in-docker container
	with_retry 5 5s testing::containerd::dind::exec " \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock image list -q"

	# Ensure that we can run some non GPU containers from within dind
	with_retry 3 5s testing::containerd::dind::exec " \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock image pull nvcr.io/xdxct/cuda:11.1.1-base-ubuntu20.04; \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock run --rm --runtime=io.containerd.runtime.v1.linux nvcr.io/xdxct/cuda:11.1.1-base-ubuntu20.04 cuda echo foo"

	# Share the volumes so that we can edit the config file and point to the new runtime
	# Share the pid so that we can ask docker to reload its config
	docker run --rm --privileged \
		--volumes-from "${containerd_dind_ctr}" \
		-v "${shared_dir}/etc/containerd/config_${version}.toml:${containerd_dind_containerd_dir}/containerd.toml" \
		--pid "container:${containerd_dind_ctr}" \
		-e RUNTIME="containerd" \
		-e RUNTIME_ARGS="--config=${containerd_dind_containerd_dir}/containerd.toml --socket=${containerd_dind_containerd_dir}/containerd.sock" \
		--name "${containerd_test_ctr}" \
		"${toolkit_container_image}" "/usr/local/xdxct" "--no-daemon"

	# We run ctr image list to ensure that containerd has successfully started in the docker-in-docker container
	with_retry 5 5s testing::containerd::dind::exec " \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock image list -q"

	# Ensure that we haven't broken non GPU containers
	with_retry 3 5s testing::containerd::dind::exec " \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock image pull nvcr.io/xdxct/cuda:11.1.1-base-ubuntu20.04; \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock run --rm --runtime=io.containerd.runtime.v1.linux nvcr.io/xdxct/cuda:11.1.1-base-ubuntu20.04 cuda echo foo"
}

# This test runs containerd setup and containerd cleanup in succession to ensure that the
# config is restored correctly.
testing::containerd::toolkit::test_config() {
	local version=${1}

	# We run ctr image list to ensure that containerd has successfully started in the docker-in-docker container
	with_retry 5 5s testing::containerd::dind::exec " \
		ctr --address=${containerd_dind_containerd_dir}/containerd.sock image list -q"

	local input_config="${shared_dir}/etc/containerd/config_${version}.toml"
	local output_config="${shared_dir}/output/config_${version}.toml"
	local output_dir=$(dirname ${output_config})

	mkdir -p ${output_dir}
	cp -p "${input_config}" "${output_config}"

	docker run --rm --privileged \
		--volumes-from "${containerd_dind_ctr}" \
		-v "${output_dir}:${output_dir}" \
		--name "${containerd_test_ctr}" \
		--entrypoint sh \
		"${toolkit_container_image}" -c "containerd setup \
			--config=${output_config} \
			--socket=${containerd_dind_containerd_dir}/containerd.sock \
			--restart-mode=none \
				/usr/local/xdxct/toolkit"

	# As a basic test we check that the config has changed
	diff "${input_config}" "${output_config}" || test ${?} -ne 0
	grep -q -E "^version = \d" "${output_config}"
	grep -q -E "default_runtime_name = \"xdxct\"" "${output_config}"

	docker run --rm --privileged \
		--volumes-from "${containerd_dind_ctr}" \
		-v "${output_dir}:${output_dir}" \
		--name "${containerd_test_ctr}" \
		--entrypoint sh \
		"${toolkit_container_image}" -c "containerd cleanup \
					--config=${output_config} \
			--socket=${containerd_dind_containerd_dir}/containerd.sock \
			--restart-mode=none \
				/usr/local/xdxct/toolkit"

	if [[ -s "${input_config}" ]]; then
		# Compare the input and output config. These should be the same.
		diff "${input_config}" "${output_config}" || true
	else
		# If the input config is empty, the output should not exist.
		test ! -e "${output_config}"
	fi
}

testing::containerd::main() {
	testing::containerd::dind::setup

	testing::containerd::toolkit::test_config empty
	testing::containerd::toolkit::test_config v1
	testing::containerd::toolkit::test_config v2

	testing::containerd::cleanup

	testing::containerd::dind::setup
	testing::containerd::toolkit::run empty
	testing::containerd::cleanup

	testing::containerd::dind::setup
	testing::containerd::toolkit::run v1
	testing::containerd::cleanup

	testing::containerd::dind::setup
	testing::containerd::toolkit::run v2
	testing::containerd::cleanup
}

testing::containerd::cleanup() {
	docker kill "${containerd_dind_ctr}" &> /dev/null || true
	docker kill "${containerd_test_ctr}" &> /dev/null || true
}
