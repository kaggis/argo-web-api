Name: ar-web-api
Summary: A/R API
Version: 1.2.1
Release: 2%{?dist}
License: ASL 2.0
Buildroot: %{_tmppath}/%{name}-buildroot
Group:     EGI/SA4
Source0: %{name}-%{version}.tar.gz
BuildRequires: golang
BuildRequires: bzr
BuildRequires: git
Requires: mongo-10gen
Requires: mongo-10gen-server
ExcludeArch: i386

%description
Installs the A/R API.

%prep
%setup

%build
export GOPATH=$PWD
cd src/api
go get api/...
cd main
go build

%install
%{__rm} -rf %{buildroot}
install --directory %{buildroot}/var/www/go-api
install --mode 755 bin/main %{buildroot}/var/www/go-api/go-api

install --directory %{buildroot}/etc/init.d
install --mode 755 go-api.init %{buildroot}/etc/init.d/

install --directory %{buildroot}/etc/init
install --mode 644 go-api.conf %{buildroot}/etc/init/

%clean
%{__rm} -rf %{buildroot}
export GOPATH=$PWD
cd src/api/main
go clean

%files
%defattr(0644,root,root)
%attr(0750,root,root) /var/www/go-api
%attr(0755,root,root) /var/www/go-api/go-api
%attr(0755,root,root) /etc/init.d/go-api.init
%attr(0644,root,root) /etc/init/go-api.conf

%changelog
* Tue Mar 25 2014 Nikos Triantafyllidis <ntrianta@grid.auth.gr> - 1.2.1-2%{?dist}
- Fixed recalculation history bug
* Tue Mar 20 2014 Nikos Triantafyllidis <ntrianta@grid.auth.gr> - 1.2.1-1%{?dist}
- Changes in results querying to reflect new database schema
* Wed Mar 19 2014 Nikolaos Triantafyllidis <ntrianta@grid.auth.gr> - 1.2.0-1%{?dist}
- Support for VOs. Changes in grouping
* Tue Mar 4 2014 Paschalis Korosoglou <pkoro@grid.auth.gr> - 1.1.1-1%{?dist}
- Suport for https
* Thu Feb 6 2014 Paschalis Korosoglou <pkoro@grid.auth.gr> - 1.1.0-2%{?dist}
- Fix in spec file
* Thu Feb 6 2014 Paschalis Korosoglou <pkoro@grid.auth.gr> - 1.1.0-1%{?dist}
- Fix in Av computation
* Thu Nov 7 2013 Paschalis Korosoglou <pkoro@grid.auth.gr> - 1.0.17-2%{?dist}
- Initial koji import