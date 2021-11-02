# certificate

## 项目说明
- 检测根目录的所有crt证书，安装到受信任的根证书颁发机构

## 环境
- Go v1.17以上
- Mingw x86_64
- upx.exe 放到GoBin （可选）

## 编译
- go build -ldflags "-s -w -H windowsgui"
- upx -9 .\certificate.exe （可选）

## 使用
- 目录结构中有certificate.exe、certmgr.exe、*.crt
- 执行certificate.exe