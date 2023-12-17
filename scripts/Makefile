ROOT = $(shell pwd)

# pm2 start ./etcd -- --config-file=./etcd.yaml
# aws 那台服务器无法加载配置文件启动
etcd:
	#scp $(ROOT)/scripts/etcd/linux_install.sh im:/data100g/t3im/etcd/
	scp $(ROOT)/scripts/etcd/etcd_alpha.yaml im:/data100g/t3im/etcd/etcd.yaml

livego:
	cd $(ROOT)/app/livego && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o livego_linux main.go
	scp $(ROOT)/app/livego/livego_linux im:/data100g/t3im/livego
	scp $(ROOT)/app/livego/livego.yaml im:/data100g/t3im/livego

# pm2 stop comet -> pm2 start comet
comet:
	cd $(ROOT)/app/comet/cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o comet main.go
	scp $(ROOT)/app/comet/cmd/comet im:/data100g/t3im/comet/
	scp $(ROOT)/app/comet/cmd/comet_alpha.toml im:/data100g/t3im/comet/comet.toml
	scp $(ROOT)/app/comet/cmd/app_alpha.yaml im:/data100g/t3im/comet/app.yaml

# pm2 stop job -> pm2 start job
job:
	cd $(ROOT)/app/job/cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o job main.go
	scp $(ROOT)/app/job/cmd/job im:/data100g/t3im/job/
	scp $(ROOT)/app/job/cmd/job_alpha.toml im:/data100g/t3im/job/job.toml
	scp $(ROOT)/app/job/cmd/app_alpha.yaml im:/data100g/t3im/job/app.yaml

# pm2 stop pionturn -> pm2 start pionturn
turn:
	cd $(ROOT)/app/pionturn/cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pionturn main.go
	scp $(ROOT)/app/pionturn/cmd/pionturn im:/data100g/t3im/pionturn
	scp $(ROOT)/app/pionturn/cmd/pionturn_alpha.toml im:/data100g/t3im/pionturn.toml

# cd /data100g/t3im/videobiz -> pm2 stop videobiz -> pm2 start videobiz
videobiz:
	cd $(ROOT)/app/videobiz && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o videobiz main.go
	scp $(ROOT)/app/videobiz/videobiz im:/data100g/t3im/videobiz/
	scp $(ROOT)/app/videobiz/config/app.yaml im:/data100g/t3im/videobiz/
	scp $(ROOT)/app/videobiz/config/database.yaml im:/data100g/t3im/videobiz
	scp $(ROOT)/app/videobiz/config/redis.yaml im:/data100g/t3im/videobiz

# cd /data100g/t3im/imbiz -> pm2 stop imbiz -> pm2 start imbiz
imbiz:
	cd $(ROOT)/app/imbiz && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o imbiz main.go
	scp $(ROOT)/app/imbiz/imbiz im:/data100g/t3im/imbiz/
	scp $(ROOT)/app/imbiz/app_alpha.yaml im:/data100g/t3im/imbiz/app.yaml

