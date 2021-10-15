##############################################################
# http://baike.corp.taobao.com/index.php/%E6%B7%98%E5%AE%9Drpm%E6%89%93%E5%8C%85%E8%A7%84%E8%8C%83 #
# http://www.rpm.org/max-rpm/ch-rpm-inside.html              #
##############################################################
Name: t-polardb-sms-agent
Version: 1.0.0
Release: %(echo $RELEASE)
# if you want use the parameter of rpm_create on build time,
# uncomment below
Summary: Please write somethings about the package here.
Group: alibaba/application
License: Commercial
AutoReqProv: none
%define _prefix /home/a/project/t-polardb-sms-agent

BuildArch:noarch

# uncomment below, if depend on other packages

#Requires: package_name = 1.0.0


%description
# if you want publish current svn URL or Revision use these macros
请你在这里描述一下关于此包的信息,并在上面的Summary后面用英文描述一下。

# support debuginfo package, to reduce runtime package size

# prepare your files
%install
# OLDPWD is the dir of rpm_create running
# _prefix is an inner var of rpmbuild,
# can set by rpm_create, default is "/home/a"
# _lib is an inner var, maybe "lib" or "lib64" depend on OS

# create dirs
mkdir -p $RPM_BUILD_ROOT%{_prefix}
cd $OLDPWD
if [ -f pom.xml ];then
  cd ..
  mvn package
  cp package/*.war $RPM_BUILD_ROOT%{_prefix}
else
    rsync -avz ../ $RPM_BUILD_ROOT%{_prefix}  --exclude=.* --exclude=rpm/
fi
# create a crontab of the package
#echo "
#* * * * * root /home/a/bin/every_min
#3 * * * * ads /home/a/bin/every_hour
#" > %{_crontab}

# package infomation
%files
# set file attribute here
%defattr(-,root,root)
# need not list every file here, keep it as this
%{_prefix}
## create an empy dir

# %dir %{_prefix}/var/log

## need bakup old config file, so indicate here

# %config %{_prefix}/etc/sample.conf

## or need keep old config file, so indicate with "noreplace"

# %config(noreplace) %{_prefix}/etc/sample.conf

## indicate the dir for crontab

# %attr(644,root,root) %{_crondir}/*

%post
#define the scripts for post install
%postun
#define the scripts for post uninstall

%changelog
* Mon Dec 7 2020 yongjie.syj
- add spec of t-polardb-sms-agent