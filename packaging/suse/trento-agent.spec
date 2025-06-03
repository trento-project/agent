#
# spec file for package trento
#
# Copyright (c) 2023 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative   .

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


Name:           trento-agent
# Version will be processed via set_version source service
Version:        0
Release:        0
License:        Apache-2.0
Summary:        Agent for Trento, the monitoring system for SAP Applications.
Group:          System/Monitoring
URL:            https://github.com/trento-project/agent
Source:         %{name}-%{version}.tar.gz
Source1:        vendor.tar.gz
ExclusiveArch:  aarch64 x86_64 s390x
BuildRoot:      %{_tmppath}/%{name}-%{version}-build
BuildRequires:  golang(API) = 1.23
# Prometheus Node Exporter is not available in SLES 16
Recommends:     golang-github-prometheus-node_exporter
Provides:       %{name} = %{version}-%{release}
Provides:       trento = %{version}-%{release}
Provides:       trento-premium = %{version}-%{release}
Obsoletes:      trento < %{version}-%{release}
Obsoletes:      trento-premium < 0.9.1-0

%description
Trento is an open cloud-native web application for SAP Applications administrators.

Trento agents are client-side processes responsible for the automatic discovery of all the components of highly available SAP Applications.

%prep
%setup -q            # unpack project sources
%setup -q -T -D -a 1 # unpack go dependencies in vendor.tar.gz, which was prepared by the source services

%define binaryname trento-agent
%define shortname agent

%build
VERSION=%{version} INSTALLATIONSOURCE=Suse BUILD_OUTPUT="./trento-agent" make build

%install

# Install the binary.
install -D -m 0755 %{binaryname} "%{buildroot}%{_bindir}/%{binaryname}"

# Install the systemd unit
install -D -m 0644 packaging/systemd/trento-agent.service %{buildroot}%{_unitdir}/trento-agent.service

# Install the default configuration files
%if 0%{?suse_version} > 1500
install -D -m 0600 packaging/config/agent.yaml %{buildroot}%{_distconfdir}/trento/agent.yaml
install -d -m 0640 %{buildroot}%{_distconfdir}/trento/plugins
%else
install -D -m 0600 packaging/config/agent.yaml %{buildroot}%{_sysconfdir}/trento/agent.yaml
install -d -m 0640 %{buildroot}%{_sysconfdir}/trento/plugins

%endif

%pre
%service_add_pre trento-agent.service

%post
%service_add_post trento-agent.service

%preun
%service_del_preun trento-agent.service

%postun
%service_del_postun trento-agent.service

%files
%defattr(-,root,root)
%doc *.md
%doc docs/*.md
%license LICENSE
%{_bindir}/%{binaryname}
%{_unitdir}/%{binaryname}.service

%if 0%{?suse_version} > 1500
%dir %_distconfdir/trento
%dir %_distconfdir/trento/plugins
%_distconfdir/trento/agent.yaml
%else
%dir %{_sysconfdir}/trento
%dir %{_sysconfdir}/trento/plugins
%config (noreplace) %{_sysconfdir}/trento/agent.yaml
%endif

%changelog
