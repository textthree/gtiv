#!/bin/bash

CURRENT_FILE_PATH=$(readlink -f "$0")
ROOT_PATH=$(dirname "$(dirname "$CURRENT_FILE_PATH")")

SVR=$1

# 创建输出目录
directory="$ROOT_PATH/dist"
if [ ! -d "$directory" ]; then
  echo "Dist directory does not exist. Creating..."
  # 创建目录
  mkdir -p "$directory"
  if [ $? -eq 0 ]; then
      echo "Dist directory created successfully."
  else
      echo "Failed to create directory."
      exit 1
  fi
fi

# etcd
copy_etcd() {
  mkdir -p $ROOT_PATH/dist/etcd
  cp $ROOT_PATH/scripts/etcd/etcd $ROOT_PATH/dist/etcd
  cp $ROOT_PATH/scripts/etcd/etcd.yaml $ROOT_PATH/dist/etcd
}

# ffmpeg
copy_ffmpeg() {
  mkdir -p $ROOT_PATH/dist/videobiz
  cp $ROOT_PATH/kit/ffmpeg/ffmpeg $ROOT_PATH/dist/videobiz
  cp $ROOT_PATH/kit/ffmpeg/ffprobe $ROOT_PATH/dist/videobiz
  cp $ROOT_PATH/kit/ffmpeg/ffmpeg $ROOT_PATH/app/videobiz
  cp $ROOT_PATH/kit/ffmpeg/ffprobe $ROOT_PATH/app/videobiz
}

# comet
build_comet() {
  mkdir -p $ROOT_PATH/dist/comet
  go build -o $ROOT_PATH/dist/comet/comet $ROOT_PATH/app/comet/cmd/main.go
  cp $ROOT_PATH/app/comet/cmd/app.yaml $ROOT_PATH/dist/comet/
  cp $ROOT_PATH/app/comet/cmd/comet.toml $ROOT_PATH/dist/comet/
}

# imbiz
build_imbiz() {
  mkdir -p $ROOT_PATH/dist/imbiz
  go build -o $ROOT_PATH/dist/imbiz/imbiz $ROOT_PATH/app/imbiz/main.go
  rsync -av --exclude='local' $ROOT_PATH/app/imbiz/config/* $ROOT_PATH/dist/imbiz/
}

# job
build_job() {
  mkdir -p $ROOT_PATH/dist/job
  go build -o $ROOT_PATH/dist/job/job $ROOT_PATH/app/job/cmd/main.go
  cp $ROOT_PATH/app/job/cmd/app.yaml $ROOT_PATH/dist/job/
  cp $ROOT_PATH/app/job/cmd/job.toml $ROOT_PATH/dist/job/
}

# livego
build_livego(){
  cd $ROOT_PATH/app/livego && go mod tidy && cd $ROOT_PATH
  mkdir -p $ROOT_PATH/dist/livego
  cd $ROOT_PATH/app/livego && go build -o $ROOT_PATH/dist/livego/livego $ROOT_PATH/app/livego/main.go
  cp $ROOT_PATH/app/livego/livego.yaml $ROOT_PATH/dist/livego/
}

# webrtc
build_webrtc(){
  mkdir -p $ROOT_PATH/dist/webrtc
  go build -o $ROOT_PATH/dist/webrtc/webrtc $ROOT_PATH/app/webrtc/cmd/main.go
  cp $ROOT_PATH/app/webrtc/cmd/config.toml $ROOT_PATH/dist/webrtc/
}

# videobiz
build_videobiz(){
  mkdir -p $ROOT_PATH/dist/videobiz
  go build -o $ROOT_PATH/dist/videobiz/videobiz $ROOT_PATH/app/videobiz/main.go
  rsync -av --exclude='local' $ROOT_PATH/app/videobiz/config/* $ROOT_PATH/dist/videobiz/
}

if [ "$SVR" = etcd ]; then
  copy_etcd
elif [ "$SVR" = comet ]; then
  build_comet
elif [ "$SVR" = imbiz ]; then
  build_imbiz
elif [ "$SVR" = job ]; then
  build_job
elif [ "$SVR" = livego ]; then
  build_livego
elif [ "$SVR" = webrtc ]; then
  build_webrtc
elif [ "$SVR" = videobiz ]; then
  build_videobiz
elif [ "$SVR" = ffmpeg ]; then
  copy_ffmpeg
fi
