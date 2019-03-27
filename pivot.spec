%define debug_package %{nil}

Name:           pivot
Version:        0.0.4
Release:        1%{?dist}
Summary:        allows moving from one OSTree deployment to another

License:        ASL 2.0
URL:            https://github.com/openshift/pivot
Source0:        https://github.com/openshift/%{name}/archive/v%{version}.tar.gz



BuildRequires:  git
BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang >= 1.6.2}
Requires:       rpm-ostree>=2019.3

%description
pivot provides a simple command allowing you to move from one OSTree
deployment to another with minimal effort.

%prep
%autosetup -n %{name}-%{version}
mkdir -p src/github.com/openshift/%{name}/
cp -rf cmd  Gopkg.lock  Gopkg.toml  LICENSE  main.go  Makefile  pivot.spec  README.md  types  utils vendor VERSION systemd src/github.com/openshift/%{name}

%build
export GOPATH=`pwd`
cd src/github.com/openshift/%{name}/
make build

%install
cd src/github.com/openshift/%{name}/
make install DESTDIR=%{buildroot}

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}
%{_prefix}/lib/systemd/system/pivot.*

%changelog
* Wed Mar 27 2019 Steve Milner <smilner@redhat.com> - 0.0.4-1
- Add basic kernel tuning functionality
- Fix previous release bump
- Don't pivot if identical sha256
- vendor: Add github.com/containers/image/docker/reference
- travis: Run tests
- tests: Make TestRunExt predictable and fast
- root: Fix missing format specifier
- Use kubelet auth if available
- README.md: Add some more concrete details on how it works
- service: Use Type=oneshot
- README.md: Add some more links for testing pivot
- README.md: Fix reboot-needed path
- tests: Set GOCACHE=off

* Mon Mar 11 2019 Yu Qi Zhang <jerzhang@redhat.com> - 0.0.3-4
- Don't pivot if identical sha256

* Tue Mar 05 2019 Yu Qi Zhang <jerzhang@redhat.com> - 0.0.3-3
- Use kubelet auth if available

* Wed Feb 20 2019 Yu Qi Zhang <jerzhang@redhat.com> - 0.0.3-2
- service: Use Type=oneshot
- README.md updates

* Tue Feb 05 2019 Jonathan Lebon <jlebon@redhat.com> - 0.0.3-1
- Add systemd service unit

* Wed Nov 14 2018 Steve Milner <smilner@redhat.com> - 0.0.2-0.1
- Makefile: Add changelog target
- cmd/root: Print pivoting msg before pull
- cmd/root: Only support the latest labels
- Support com.coreos.ostree-commit too
- cmd/root: Make OSTree version we pivot to more prominent
- cmd/root: Don't rely on refs, use checksum from label
- cmd: only create, don't run the container
- cmd: add --unchanged-exit-77
- cmd: fix capitalization
- Namespace move from ashcrow to openshift
- README.md: update with more information
- cmd: drop networking in the container
- gofmt: Update formatting
- cmd: ask rpm-ostree for state instead of marker file
- cmd: implement full idempotency
- types: use permalink to skopeo spec
- utils: use glog.Infof for consistency
- cmd: add the default cmdline flags
- cmd: drop --touch-if-changed
- Use pivot:// prefix for rpm-ostree custom URL
- Pass custom origin information to rpm-ostree
- Canonicalize to name@digest
- Merge pull request #10 from ashcrow/misc-enhancements
- utils: Add basic unittests
- main.go: Generate header from build info
- .travis.yml: Switch to 1.1[0,1].x golang versions
- .gitignore: Ignore .vscode


* Thu Aug  9 2018 Steve Milner <smilner@redhat.com> - 0.0.1-0.1
- Multiple fixes

* Wed Jun 13 2018 Steve Milner <smilner@redhat.com> - 0.0.0-0.1
- Initial spec
