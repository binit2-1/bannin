# Build with `make rpm`. The tarball includes vendor/, so the build needs no network.
%global plasmoid_id io.github.binit2-1.bannin

Name:           bannin
Version:        %{?pkg_version}%{!?pkg_version:0.1.0}
Release:        1%{?dist}
Summary:        Memory watchman for KDE Plasma that warns before your system freezes
# TODO: choose a license and add a LICENSE file before publishing.
License:        LicenseRef-TBD
URL:            https://github.com/binit2-1/bannin
Source0:        %{name}-%{version}.tar.gz

ExclusiveArch:  %{golang_arches}
BuildRequires:  golang >= 1.26
BuildRequires:  systemd-rpm-macros
Requires:       plasma-workspace
Recommends:     systemd-oomd-defaults

%description
Bannin watches RAM, swap and memory pressure. A panel widget shows the level
at a glance; a small user service warns with a notification before the
system runs out of memory, and can stop dev tools (Metro, Gradle, the Android
emulator) left running after their terminal closed. Auto-stop starts in dry
run, logging what it would have stopped.

%prep
%autosetup

%build
export GOFLAGS="-mod=vendor -trimpath -buildmode=pie"
export CGO_ENABLED=0
for cmd in bannin bannind; do
    go build -ldflags "-s -w -X main.version=%{version}" -o bin/$cmd ./cmd/$cmd
done

%check
GOFLAGS="-mod=vendor" go test ./...

%install
install -Dm755 bin/bannin  %{buildroot}%{_bindir}/bannin
install -Dm755 bin/bannind %{buildroot}%{_bindir}/bannind
install -Dm644 dist/bannind.service %{buildroot}%{_userunitdir}/bannind.service
sed -i 's|^ExecStart=.*|ExecStart=%{_bindir}/bannind|' %{buildroot}%{_userunitdir}/bannind.service
mkdir -p %{buildroot}%{_datadir}/plasma/plasmoids/%{plasmoid_id}
cp -r ui/plasma/. %{buildroot}%{_datadir}/plasma/plasmoids/%{plasmoid_id}/

%post
%systemd_user_post bannind.service

%preun
%systemd_user_preun bannind.service

%files
%doc README.md
%{_bindir}/bannin
%{_bindir}/bannind
%{_userunitdir}/bannind.service
%{_datadir}/plasma/plasmoids/%{plasmoid_id}/
