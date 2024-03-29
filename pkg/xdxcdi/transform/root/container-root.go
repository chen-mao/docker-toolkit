package root

import (
	"fmt"
	"strings"

	"tags.cncf.io/container-device-interface/specs-go"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/xdxcdi/transform"
)

// containerRootTransformer transforms the roots of container paths in a CDI spec.
type containerRootTransformer transformer

var _ transform.Transformer = (*containerRootTransformer)(nil)

// Transform replaces the root in a spec with a new root.
// It walks the spec and replaces all container paths that start with root with the target root.
func (t containerRootTransformer) Transform(spec *specs.Spec) error {
	if spec == nil {
		return nil
	}

	for _, d := range spec.Devices {
		d := d
		if err := t.applyToEdits(&d.ContainerEdits); err != nil {
			return fmt.Errorf("failed to apply root transform to device %s: %w", d.Name, err)
		}
	}

	if err := t.applyToEdits(&spec.ContainerEdits); err != nil {
		return fmt.Errorf("failed to apply root transform to spec: %w", err)
	}
	return nil
}

func (t containerRootTransformer) applyToEdits(edits *specs.ContainerEdits) error {
	for i, dn := range edits.DeviceNodes {
		edits.DeviceNodes[i] = t.transformDeviceNode(dn)
	}

	for i, hook := range edits.Hooks {
		edits.Hooks[i] = t.transformHook(hook)
	}

	for i, mount := range edits.Mounts {
		edits.Mounts[i] = t.transformMount(mount)
	}

	return nil
}

func (t containerRootTransformer) transformDeviceNode(dn *specs.DeviceNode) *specs.DeviceNode {
	dn.Path = t.transformPath(dn.Path)

	return dn
}

func (t containerRootTransformer) transformHook(hook *specs.Hook) *specs.Hook {
	// The Path in the startContainer hook MUST resolve in the container namespace.
	if hook.HookName == "startContainer" {
		hook.Path = t.transformPath(hook.Path)
	}

	// The createContainer and startContainer hooks MUST execute in the container namespace.
	if hook.HookName != "createContainer" && hook.HookName != "startContainer" {
		return hook
	}

	var args []string
	for _, arg := range hook.Args {
		if !strings.Contains(arg, "::") {
			args = append(args, t.transformPath(arg))
			continue
		}

		// For the 'create-symlinks' hook, special care is taken for the
		// '--link' flag argument which takes the form <target>::<link>.
		// Both paths, the target and link paths, are transformed.
		split := strings.SplitN(arg, "::", 2)
		split[0] = t.transformPath(split[0])
		split[1] = t.transformPath(split[1])
		args = append(args, strings.Join(split, "::"))
	}
	hook.Args = args

	return hook
}

func (t containerRootTransformer) transformMount(mount *specs.Mount) *specs.Mount {
	mount.ContainerPath = t.transformPath(mount.ContainerPath)
	return mount
}

func (t containerRootTransformer) transformPath(path string) string {
	return (transformer)(t).transformPath(path)
}
