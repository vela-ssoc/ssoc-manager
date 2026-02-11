# SSOCv2


### ssoc-manager.service

```text
[Unit]
Description=SSOCv2 MANAGER
After=network-online.target
Wants=network-online.target
Documentation=https://github.com/vela-ssoc

[Service]
Type=simple
WorkingDirectory=/vdb/ssoc/manager
ExecStart=/vdb/ssoc/manager/ssoc-manager
KillSignal=SIGINT
TimeoutStopSec=10
Restart=on-failure
RestartSec=5
Environment=TERM=xterm-256color
StandardOutput=append:/vdb/ssoc/manager/resources/log/systemd-stdout.log
StandardError=append:/vdb/ssoc/manager/resources/log/systemd-stderr.log

[Install]
WantedBy=multi-user.target
```
