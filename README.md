# WSystemd

WSystemd 是一个分布式任务调度和进程管理系统，支持单机和集群模式运行。它提供了强大的任务生命周期管理、自动故障恢复和负载均衡能力。

## ✨ 特性

- 🚀 支持单机/集群模式运行
- 🔄 基于 etcd 的服务注册与发现
- ⚖️ 智能任务负载均衡
- 🔄 进程生命周期管理
- 🛡️ 任务健康检查和自动恢复
- 🌐 RESTful API 接口
- 📊 实时任务状态监控

## 🏗️ 系统架构
```
┌─────────────┐
│    etcd     │      服务注册与发现
└─────────────┘
      ▲
      │
      ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Worker    │◄──►│   Worker    │◄──►│   Worker    │  任务调度层
└─────────────┘    └─────────────┘    └─────────────┘
      ▲                  ▲                  ▲
      │                  │                  │
      ▼                  ▼                  ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Tasks     │    │   Tasks     │    │   Tasks     │  任务执行层
└─────────────┘    └─────────────┘    └─────────────┘
```

## 🚀 快速开始

### 环境要求

- Go 1.23+
- MySQL 5.7+
- etcd 3.5+ (集群模式需要)

### 📦 安装
```bash
# 克隆项目
git clone https://github.com/yourusername/wsystemd.git

# 进入项目目录
cd wsystemd

# 安装依赖
go mod tidy

# 编译
go build
```

### ⚙️ 配置
创建配置文件 `./conf/vsystemd.yaml`:
```yaml
dev:
  mysql:
    wsystemd:
      host: 172.16.27.66
      port: 3306
      user: will
      password: "admin@123"
      dataBase: wsystemd
      parseTime: true
      maxIdleConns: 10
      maxOpenConns: 30
      connMaxLifetime: 28800
      connMaxIdletime: 7200
  scheduleTimeTicker: 30
  # 是否为单机模式, 集群模式需要配置 etcd
  singleMode: true
  etcd: 172.16.27.66
```

### 📝 初始化数据库
```bash
mysql -u root -p < sql/task.sql
```

### 🚀 运行
```bash
./wsystemd --server-port 8500
```

## 📡 API 接口

### 创建任务
```http
POST /v1/jobs/submit

{
    "node": "",
    "dc": "",
    "ip": "",
    "loadMethod": "",
    "doOnce": false,
    "run": {
        "cmd": "/path/to/your/app",
        "args": ["arg1", "arg2"],
        "outfile": "./log/out.log",
        "errfile": "./log/err.log",
        "type": "cmdline"
    }
}
```

### 停止任务
```http
PUT /v1/jobs/{jobId}/stop
```

### 任务心跳上报
```http
POST /v1/agent/tasks/report?token={主机名称}:{jobId}
```

## 🛠️ 核心功能

### 进程管理
- 完整的进程生命周期管理
- 自动重启机制
- 资源限制和监控

## 🗺️ 开发计划

- [ ] 支持更多任务调度策略
- [ ] 添加 Web 管理界面
- [ ] 支持任务依赖关系
- [ ] 添加任务执行统计
- [ ] 优化性能监控, 任务状态监控
- [ ] 支持容器化部署
- [ ] 完善监控告警机制, 资源使用告警
- [ ] 集群模式下的节点任务故障转移
- [ ] 丰富集群模式下的任务调度策略

## 🤝 贡献指南
欢迎提交 Issue 和 Pull Request！

联系方式：<826895143@qq.com>

## 📄 许可证
[MIT License](LICENSE)


