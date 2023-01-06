# proxy

support http proxy and vless over h2

自用的一个工具，根据自己需求开发，集成静态 http 资源 server、http porxy server、vless server 和 docker registry。

## 参考配置文件

```yaml
log: "info" # 日志级别
tls: # 证书
  cert: "cert.pem"
  key: "key.pem"
http: ":80"
https: ":443"
enable-http3: true # 启用 http3/quic
users: # 用户定义
  - username: admin # 可用于 registry、vless、proxy 认证
    password: password
    uuid: bea952e3-eeeb-40b6-9fc1-9237114c0d5f # 可用于 vless 认证 
component: # 组件定义
  registry: # registry 相关定义 
    enable: true
    config: # registry 配置文件，参考 https://docs.docker.com/registry/configuration/
      version: 0.1
      storage:
        filesystem:
          rootdirectory: /var/lib/registry
          maxthreads: 100
        delete:
          enabled: true
        redirect:
          disable: false
        cache:
          blobdescriptor: inmemory
          blobdescriptorsize: 10000
        maintenance:
          uploadpurging:
            enabled: true
            age: 168h
            interval: 24h
            dryrun: false
      auth:
        global_user:
          realm: "registry"
      http:
        secret: secrets
        relativeurls: true
        draintimeout: 60s
  vless: # vless 相关定义
    path: "/vless/:username/:password" # vless over h2 path，采用 gin path 参数，支持 传入用户名、密码
    enable: true
    auth: "global_user"
  proxy: # http proxy 配置
    enable: true
    auth: "global_user"
    active: true
  static: # 静态 http 配置
    enable: true
    root: /srv/lib/html

```
