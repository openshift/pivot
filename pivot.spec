%define debug_package %{nil}

Name:           pivot
Version:        0.0.1
Release:        0.1%{?dist}
Summary:        allows moving from one OSTree deployment to another

License:        ASL 2.0
URL:            https://github.com/openshift/pivot
Source0:        https://github.com/openshift/%{name}/archive/v%{version}.tar.gz



BuildRequires:  git
BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang >= 1.6.2}

%description
pivot provides a simple command allowing you to move from one OSTree
deployment to another with minimal effort.

%prep
%autosetup -n %{name}-%{version}
mkdir -p src/github.com/openshift/%{name}/
cp -rf cmd  Gopkg.lock  Gopkg.toml  LICENSE  main.go  Makefile  pivot.spec  README.md  types  utils vendor VERSION src/github.com/openshift/%{name}

%build
export GOPATH=`pwd`
cd src/github.com/openshift/%{name}/
make build

%install
ls -la
install -d %{buildroot}%{_bindir}
install --mode 755 src/github.com/openshift/%{name}/%{name} %{buildroot}%{_bindir}/%{name}

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}


%changelog
* Thu Aug  9 2018 Steve Milner <smilner@redhat.com> - 0.0.1-0.1
- Multiple fixes

* Wed Jun 13 2018 Steve Milner <smilner@redhat.com> - 0.0.0-0.1
- Initial spec
