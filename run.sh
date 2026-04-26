#!/bin/bash
# N1 打印服务编译/运行脚本

set -euo pipefail

echo "========================================"
echo "  N1 打印服务 - 编译/运行"
echo "========================================"

CONFIG_FILE="./config.json"

if [ ! -f "$CONFIG_FILE" ]; then
  echo ">>> 未找到 config.json，创建模板..."
  cp ./config.example.json "$CONFIG_FILE"
  echo "请先编辑 $CONFIG_FILE 填写真实配置，再重新运行。"
  exit 0
fi

if grep -q '"mqtt_username": "CHANGE_ME"' "$CONFIG_FILE" || \
   grep -q '"mqtt_password": "CHANGE_ME"' "$CONFIG_FILE" || \
   grep -q '"shop_slug": "CHANGE_ME"' "$CONFIG_FILE"; then
  echo "检测到 config.json 仍包含必填占位值，请先填写真实配置。"
  exit 1
fi

echo ">>> 1. 安装系统依赖..."
sudo apt-get update -qq
sudo apt-get install -y libusb-1.0-0 libusb-1.0-0-dev pkg-config udev > /dev/null 2>&1

echo ">>> 2. 配置 USB 权限..."
if ! grep -q "0483.*070b" /etc/udev/rules.d/99-printer.rules 2>/dev/null; then
    sudo tee /etc/udev/rules.d/99-printer.rules > /dev/null << 'EOF'
SUBSYSTEM=="usb", ATTR{idVendor}=="0483", ATTR{idProduct}=="070b", MODE="0666", GROUP="plugdev"
EOF
    sudo udevadm control --reload-rules
    sudo udevadm trigger
    echo "USB 规则已配置"
else
    echo "USB 规则已存在"
fi

echo ">>> 3. 下载 Go 依赖..."
go mod tidy

echo ">>> 4. 编译中..."
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o printer-arm64 .

if [ ! -f "printer-arm64" ]; then
    echo "错误: 编译失败!"
    exit 1
fi

echo "编译成功!"

echo ""
echo "输入 y 运行打印服务，其他键退出:"
read -n 1 -t 10 answer || true

echo ""
if [ "${answer:-}" = "y" ] || [ "${answer:-}" = "Y" ]; then
    echo "启动打印服务..."
    ./printer-arm64 -config "$CONFIG_FILE"
else
    echo "退出。如需运行，请执行: ./printer-arm64 -config $CONFIG_FILE"
fi
