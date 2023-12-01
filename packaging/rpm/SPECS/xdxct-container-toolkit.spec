Name: xdxct-container-toolkit
Version: %{version}
Release: %{release}
Group: Development Tools

Vendor: XDXCT CORPORATION
Packager: XDXCT CORPORATION <tools@xdxct.com>

Summary: XDXCT Container Toolkit
URL: https://github.com/chen-mao/docker-toolkit
License: Apache-2.0

Source0: xdxct-container-runtime-hook
Source1: xdxct-ctk
Source2: LICENSE
Source3: xdxct-container-runtime
Source4: xdxct-container-runtime.cdi
Source5: xdxct-container-runtime.legacy

Obsoletes: xdxct-container-runtime <= 3.5.0-1, xdxct-container-runtime-hook <= 1.4.0-2
Provides: xdxct-container-runtime
Provides: xdxct-container-runtime-hook
Requires: libxdxct-container-tools >= %{libxdxct_container_tools_version}, libxdxct-container-tools < 2.0.0
Requires: xdxct-container-toolkit-base == %{version}-%{release}

%if 0%{?suse_version}
Requires: libseccomp2
Requires: libapparmor1
%else
Requires: libseccomp
%endif

%description
Provides tools and utilities to enable GPU support in containers.

%prep
cp %{SOURCE0} %{SOURCE1} %{SOURCE2} %{SOURCE3} %{SOURCE4} %{SOURCE5} .

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 -t %{buildroot}%{_bindir} xdxct-container-runtime-hook
install -m 755 -t %{buildroot}%{_bindir} xdxct-container-runtime
install -m 755 -t %{buildroot}%{_bindir} xdxct-container-runtime.cdi
install -m 755 -t %{buildroot}%{_bindir} xdxct-container-runtime.legacy
install -m 755 -t %{buildroot}%{_bindir} xdxct-ctk

%post
if [ $1 -gt 1 ]; then  # only on package upgrade
  mkdir -p %{_localstatedir}/lib/rpm-state/xdxct-container-toolkit
  cp -af %{_bindir}/xdxct-container-runtime-hook %{_localstatedir}/lib/rpm-state/xdxct-container-toolkit
fi

%posttrans
if [ ! -e %{_bindir}/xdxct-container-runtime-hook ]; then
  # repairing lost file xdxct-container-runtime-hook
  cp -avf %{_localstatedir}/lib/rpm-state/xdxct-container-toolkit/xdxct-container-runtime-hook %{_bindir}
fi
rm -rf %{_localstatedir}/lib/rpm-state/xdxct-container-toolkit
ln -sf %{_bindir}/xdxct-container-runtime-hook %{_bindir}/xdxct-container-toolkit

# Generate the default config; If this file already exists no changes are made.
# %{_bindir}/xdxct-ctk --quiet config --config-file=%{_sysconfdir}/xdxct-container-runtime/config.toml --in-place

%postun
if [ "$1" = 0 ]; then  # package is uninstalled, not upgraded
  if [ -L %{_bindir}/xdxct-container-toolkit ]; then rm -f %{_bindir}/xdxct-container-toolkit; fi
fi

%files
%license LICENSE
%{_bindir}/xdxct-container-runtime-hook

%changelog
# As of 1.10.0-1 we generate the release information automatically
* %{release_date} XDXCT CORPORATION <tools@xdxct.com> %{version}-%{release}
- See https://gitlab.com/nvidia/container-toolkit/container-toolkit/-/blob/%{git_commit}/CHANGELOG.md
- Bump libnvidia-container dependency to libnvidia-container-tools >= %{libnvidia_container_tools_version}

# The BASE package consists of the NVIDIA Container Runtime and the NVIDIA Container Toolkit CLI.
# This allows the package to be installed on systems where no NVIDIA Container CLI is available.
%package base
Summary: XDXCT Container Toolkit Base
Obsoletes: xdxct-container-runtime <= 3.5.0-1, xdxct-container-runtime-hook <= 1.4.0-2
Provides: xdxct-container-runtime
# Since this package allows certain components of the XDXCT Container Toolkit to be installed separately
# it conflicts with older versions of the xdxct-container-toolkit package that also provide these files.
# Conflicts: xdxct-container-toolkit <= 1.10.0-1

%description base
Provides tools such as the XDXCT Container Runtime and XDXCT Container Toolkit CLI to enable GPU support in containers.

%files base
%license LICENSE
%{_bindir}/xdxct-container-runtime
%{_bindir}/xdxct-ctk

# The OPERATOR EXTENSIONS package consists of components that are required to enable GPU support in Kubernetes.
# This package is not distributed as part of the XDXCT Container Toolkit RPMs.
%package operator-extensions
Summary: XDXCT Container Toolkit Operator Extensions
Requires: xdxct-container-toolkit-base == %{version}-%{release}

%description operator-extensions
Provides tools for using the XDXCT Container Toolkit with the GPU Operator

%files operator-extensions
%license LICENSE
%{_bindir}/xdxct-container-runtime.cdi
%{_bindir}/xdxct-container-runtime.legacy