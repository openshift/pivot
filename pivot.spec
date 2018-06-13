%define debug_package %{nil}

Name:           pivot
Version:        0.0.0
Release:        0.1%{?dist}
Summary:        allows moving from one OSTree deployment to another

License:        ASL 2.0
URL:            https://github.com/ashcrow/pivot
Source0:        https://github.com/ashcrow/%{name}/archive/v%{version}.tar.gz



BuildRequires:  git
%if 0%{?fedora}
BuildRequires:  %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang >= 1.6.2}
%endif #fedora
%if 0%{?centos}
BuildRequires:  golang
%endif #centos

%description
pivot provides a simple command allowing you to move from one OSTree
deployment to another with minimal effort.

%prep
%autosetup -n %{name}-%{version}
mkdir -p src/github.com/ashcrow/%{name}/
cp -rf cmd  Gopkg.lock  Gopkg.toml  LICENSE  main.go  Makefile  pivot.spec  README.md  types  utils vendor VERSION src/github.com/ashcrow/%{name}

%build
export GOPATH=`pwd`
cd src/github.com/ashcrow/%{name}/
make build

%install
ls -la
install -d %{buildroot}%{_bindir}
install --mode 755 src/github.com/ashcrow/%{name}/%{name} %{buildroot}%{_bindir}/%{name}

%files
%license LICENSE
%doc README.md
%{_bindir}/%{name}


%changelog
* Wed Jun 13 2018 Steve Milner <smilner@redhat.com> - 0.0.0-0.1
- Initial spec
