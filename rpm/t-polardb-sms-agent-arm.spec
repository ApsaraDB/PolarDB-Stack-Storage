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

BuildArch:aarch64

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

go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE=gitlab.alibaba-inc.com

sudo chmod 700 /root/.ssh
sudo chmod 777 /root/.ssh/id_rsa
sudo chmod 777 /root/.ssh/id_rsa.pub
sudo sh -c 'cat > /root/.ssh/id_rsa <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApV1uIHglzB3rUw7O60T0nnIuR81v2mj2rOTKwX9B7PwIuCLO
ElUtefOSBRTuI2TVHHvn4vVWzmVb5xv817wbnXRuwx0UydQKA33k3IJ25xNzGoM5
xuJWvYtr7rgDkozqOsIgH5dtvIC6fupjF5HljFgAmiAP/vOcnisJXepznUfOLMuL
57oqd8ek//84rxgmQNVIleKd7wxpM65SXckQOnGDzk8D91N1N4XU3OLgk4oMa/CJ
L7GnZMJdWtuVSAGRrppXKf9QkBV4tT0YqhklYCFttNTtLXaLw8G6vZx7GQuhvQOS
tFb3CHLyOzcgqnmifBA8HTVynLi8i2Wz9VxRDQIDAQABAoIBAHJiEg77jImIGIK3
N4GdjOcca87W14voFti42VbGw789Bnr3+DrOGErGpDZHqAFrec3eFyPyOD1D9zIB
Nf+z6hYbt5HWx85jkRmhN3Ef/UcROQkZxBlB7mXlzp6tQitYtZw3MyknAYzmLhoj
3q8rB/Dv3lq92tKrh6YQdsYzqg0bxWMchXByn9rAmU27ywVHjJ6PK/lzfixwv9Zh
V8/ePkMEF9H43GpxQliZ3f7mBMKiEgpl5F7lIwREsmOvTw+7cCGuN5w6lkCGsA5S
0uiRpFlOm6z6H7jG49dyF90PHn2o2Lt3RzD8kvKrQdVEhjJ0cbo3whFtfZooNkEY
L9HsTjECgYEA0HWb5rHEmTXowWDyjT9C2O9VX0+gBA/ZVM8jgTtCIJB4CITqGQd6
PQTOCWiw8KDYNlPfff+g2sJ5W7WYdCYxcMRcgbxwT1qTsgRdsAiOK05ybgAPWrv0
uVnipAeGVusBeddr1CWMs277Hj21cyRmJFKnGljJKoZGmAuS+9KMcXMCgYEAyxPb
izSagoWFmEv2GV73piYKL6hvPSUStUxjV3yfY0/gb+Pl//LSH0WlcBmdDkf/ONmb
0HPBi/Q3fCpkQHfim2iF1v7QvztaJuQtrBuFE3E9kw5N0wgAkJOJew5TEO4A/mv/
JC8JfaJ39fYELUrVSC6NdAeDtjXm5vt7ysKJk38CgYBU67RpFH4slAOP07i7xcur
qEQ1IbDkNriojgG+wa88qN3dSpg7PgqUFQMCXj3GqR+rchuXrq2OsK7Tp3TFzFFT
yQqOZ3+xNAr6+EBaWAHiroB7Q1b99ZfKck2b2NznR7FAO3vX3rwk1M9EEIt8rpVV
sr4UQ5sf0besdPqZz7oa9QKBgFjO0k/KLVeu9IFplrH5qetq70FwM1VWBRxrz3XO
9hUENW45B7gqhGFQn8yqJti7+4zs/Qrn1FhT8H/IOhdHVj4IM5+Vx8wZNI+VcbO2
RTf/aaIJu1byROz02EaMOR9KNf0NVYKJX2klx7g0Yhc8hpEOaqW3M45XfCa/C5/b
+zYZAoGBAJu+FRspbhAWzduPf2YdEiAwTNPcZSgl3RA4ZYrseXEqlUaL77xrCnF7
vddYW9TZlE53ZdGfQYC13rQWlfUls154kpMTJIoC8cAE0+/4A7oUTiGlEsEshRYX
+KTYp2HuPAnbmD/rVhiLICgELvVnj3Ocl5LNNlgayQaaFbYdAaKk
-----END RSA PRIVATE KEY-----
EOF'
sudo sh -c 'echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQClXW4geCXMHetTDs7rRPSeci5HzW/aaPas5MrBf0Hs/Ai4Is4SVS1585IFFO4jZNUce+fi9VbOZVvnG/zXvBuddG7DHRTJ1AoDfeTcgnbnE3MagznG4la9i2vuuAOSjOo6wiAfl228gLp+6mMXkeWMWACaIA/+85yeKwld6nOdR84sy4vnuip3x6T//zivGCZA1UiV4p3vDGkzrlJdyRA6cYPOTwP3U3U3hdTc4uCTigxr8Ikvsadkwl1a25VIAZGumlcp/1CQFXi1PRiqGSVgIW201O0tdovDwbq9nHsZC6G9A5K0VvcIcvI7NyCqeaJ8EDwdNXKcuLyLZbP1XFEN root@efc-builder" > /root/.ssh/id_rsa.pub'

sudo sh -c 'echo "Host gitlab.alibaba-inc.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config'
git config --global url."git@gitlab.alibaba-inc.com:".insteadOf "https://gitlab.alibaba-inc.com/"
CodeSource=http://gitlab.alibaba-inc.com/polar-as/polardb-sms.git

CodeBranch=`git branch`

CodeVersion=`git rev-parse HEAD`

BuildDate=`date -u +"%Y-%m-%dT%H:%M:%SZ"`

# create dirs
mkdir -p $RPM_BUILD_ROOT%{_prefix}
mkdir -p $RPM_BUILD_ROOT%{_prefix}/bin
cd $OLDPWD/..
#export GOPROXY=http://11.239.161.136 GOSUMDB=off
go mod download
go build -ldflags "-X 'polardb-sms/pkg/version.GitBranch=$CodeBranch' \
                                              -X 'polardb-sms/pkg/version.GitCommit=$CodeVersion'\
                                              -X 'polardb-sms/pkg/version.BuildDate=$BuildDate' \
                                              -X 'polardb-sms/pkg/version.Module=$CodeSource'" -o polardb-sms-agent ./cmd/agent
mv polardb-sms-agent $RPM_BUILD_ROOT%{_prefix}/bin/

# package infomation
%files
# set file attribute here
%defattr(-,root,root)
# need not list every file here, keep it as this
%{_prefix}
## create an empty dir

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