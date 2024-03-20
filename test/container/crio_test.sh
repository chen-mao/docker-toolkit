testing::crio::hook_created() {
	testing::docker_run::toolkit::shell 'crio setup /run/xdxct/toolkit'

	test ! -z "$(ls -A "${shared_dir}${CRIO_HOOKS_DIR}")"

	cat "${shared_dir}${CRIO_HOOKS_DIR}/${CRIO_HOOK_FILENAME}" | \
		jq -r '.hook.path' | grep -q "/run/xdxct/toolkit/"
	test $? -eq 0
	cat "${shared_dir}${CRIO_HOOKS_DIR}/${CRIO_HOOK_FILENAME}" | \
		jq -r '.hook.env[0]' | grep -q ":/run/xdxct/toolkit"
	test $? -eq 0
}

testing::crio::hook_cleanup() {
	testing::docker_run::toolkit::shell 'crio cleanup'

	test -z "$(ls -A "${shared_dir}${CRIO_HOOKS_DIR}")"
}

testing::crio::main() {
	testing::crio::hook_created
	testing::crio::hook_cleanup
}

testing::crio::cleanup() {
	:
}
