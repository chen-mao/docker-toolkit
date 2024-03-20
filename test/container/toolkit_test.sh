testing::toolkit::install() {
	local -r uid=$(id -u)
	local -r gid=$(id -g)

	local READLINK="readlink"
	local -r platform=$(uname)
	if [[ "${platform}" == "Darwin" ]]; then
		READLINK="greadlink"
	fi

	testing::docker_run::toolkit::shell 'toolkit install --toolkit-root=/usr/local/xdxct/toolkit'
	docker run --rm -v "${shared_dir}:/work" alpine sh -c "chown -R ${uid}:${gid} /work/"

	# Ensure toolkit dir is correctly setup
	test ! -z "$(ls -A "${shared_dir}/usr/local/xdxct/toolkit")"

	test -L "${shared_dir}/usr/local/xdxct/toolkit/libxdxct-container.so.1"
	test -e "$(${READLINK} -f "${shared_dir}/usr/local/xdxct/toolkit/libxdxct-container.so.1")"
	test -L "${shared_dir}/usr/local/xdxct/toolkit/libxdxct-container-go.so.1"
	test -e "$(${READLINK} -f "${shared_dir}/usr/local/xdxct/toolkit/libxdxct-container-go.so.1")"

	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-cli"
	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime-hook"
	test -L "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-toolkit"
	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime"

	grep -q -E "xdxct driver modules are not yet loaded, invoking runc directly" "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime"
	grep -q -E "exec runc \".@\"" "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime"

	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-cli.real"
	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime-hook.real"
	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.real"

	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.experimental"
	test -e "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.experimental.real"

	grep -q -E "xdxct driver modules are not yet loaded, invoking runc directly" "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.experimental"
	grep -q -E "exec runc \".@\"" "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.experimental"
	grep -q -E "LD_LIBRARY_PATH=/run/xdxct/driver/usr/lib64:\\\$LD_LIBRARY_PATH " "${shared_dir}/usr/local/xdxct/toolkit/xdxct-container-runtime.experimental"

	test -e "${shared_dir}/usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml"

	# Ensure that the config file has the required contents.
	# NOTE: This assumes that RUN_DIR is '/run/xdxct'
	local -r xdxct_run_dir="/run/xdxct"
	grep -q -E "^\s*ldconfig = \"@${xdxct_run_dir}/driver/sbin/ldconfig(.real)?\"" "${shared_dir}/usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml"
	grep -q -E "^\s*root = \"${xdxct_run_dir}/driver\"" "${shared_dir}/usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml"
	grep -q -E "^\s*path = \"/usr/local/xdxct/toolkit/xdxct-container-cli\"" "${shared_dir}/usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml"
	grep -q -E "^\s*path = \"/usr/local/xdxct/toolkit/xdxct-ctk\"" "${shared_dir}/usr/local/xdxct/toolkit/.config/xdxct-container-runtime/config.toml"
}

testing::toolkit::delete() {
	testing::docker_run::toolkit::shell 'mkdir -p /usr/local/xdxct/delete-toolkit'
	testing::docker_run::toolkit::shell 'touch /usr/local/xdxct/delete-toolkit/test.file'
	testing::docker_run::toolkit::shell 'toolkit delete --toolkit-root=/usr/local/xdxct/delete-toolkit'

	test ! -z "$(ls -A "${shared_dir}/usr/local/xdxct")"
	test ! -e "${shared_dir}/usr/local/xdxct/delete-toolkit"
}

testing::toolkit::main() {
	testing::toolkit::install
	testing::toolkit::delete
}

testing::toolkit::cleanup() {
	:
}
