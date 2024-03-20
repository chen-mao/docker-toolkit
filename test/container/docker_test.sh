readonly docker_dind_ctr="container-config-docker-dind-ctr-name"
readonly docker_test_ctr="container-config-docker-test-ctr-name"
readonly docker_dind_socket="/run/xdxct/docker.sock"

testing::docker::dind::setup() {
	# Docker creates /etc/docker when starting
	# by default there isn't any config in this directory (even after the daemon starts)
	docker run -d --rm --privileged \
		-v "${shared_dir}/etc/docker:/etc/docker" \
		-v "${shared_dir}/run/xdxct:/run/xdxct" \
		-v "${shared_dir}/usr/local/xdxct:/usr/local/xdxct" \
		--name "${docker_dind_ctr}" \
		docker:stable-dind -H unix://${docker_dind_socket}
}

testing::docker::dind::exec() {
	docker exec "${docker_dind_ctr}" sh -c "$*"
}

testing::docker::toolkit::run() {
	# Share the volumes so that we can edit the config file and point to the new runtime
	# Share the pid so that we can ask docker to reload its config
	docker run -d --rm --privileged \
		--volumes-from "${docker_dind_ctr}" \
		--pid "container:${docker_dind_ctr}" \
		-e RUNTIME_ARGS="--socket ${docker_dind_socket}" \
		--name "${docker_test_ctr}" \
		"${toolkit_container_image}" "/usr/local/xdxct" "--no-daemon"

	# Ensure that we haven't broken non GPU containers
	with_retry 3 5s testing::docker::dind::exec docker run -t alpine echo foo
}

testing::docker::main() {
	testing::docker::dind::setup
	testing::docker::toolkit::run
}

testing::docker::cleanup() {
	docker kill "${docker_dind_ctr}" &> /dev/null || true
	docker kill "${docker_test_ctr}" &> /dev/null || true
}
