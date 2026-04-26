# N1 打印服务

独立运行的点餐打印服务，专为 N1 盒子设计。

## 推荐方案

优先使用当前 Go 版本。

- 依赖更少
- 运行更稳
- 适合 Debian 12 / Armbian 长期部署
- 配置统一走 `config.json`，也支持环境变量覆盖
- 支持启动时自动从后端同步最新 `mqtt_secret`

根目录的 `mqtt.py` 仅保留为 legacy 备用脚本，不建议作为长期生产方案。

## 目录说明

- `main.go` / `service.go` / `usb_linux.go`: 主程序
- `config.go`: 配置加载和校验
- `config.example.json`: 配置模板
- `run.sh`: 本地编译和手动运行辅助脚本
- `printer-arm64`: 编译产物，不建议提交到仓库

## 安装依赖

```bash
sudo apt-get update
sudo apt-get install -y libusb-1.0-0 libusb-1.0-0-dev pkg-config udev
```

## USB 权限

```bash
sudo cat > /etc/udev/rules.d/99-printer.rules << 'EOF'
SUBSYSTEM=="usb", ATTR{idVendor}=="0483", ATTR{idProduct}=="070b", MODE="0666", GROUP="plugdev"
EOF
sudo udevadm control --reload-rules
sudo udevadm trigger
sudo modprobe -r usblp || true
```

## 配置

先复制模板：

```bash
cp config.example.json config.json
```

然后修改这些字段：

- `config_sync_base_url`
- `mqtt_username`
- `mqtt_password`
- `shop_slug`
- `printer_vendor`
- `printer_product`

说明：
- `config_sync_base_url` 推荐填 `https://food2api.serbia70.com`
- `mqtt_secret` 可以留空。程序启动时会按 `shop_slug` 自动向后端拉最新 secret；拉取成功后会自动写回本地 `config.json`
- `mqtt_order_topic` 留空即可，程序会自动按 `restaurant/<shop_slug>/<mqtt_secret>/order` 生成
- 环境变量会覆盖 `config.json`

## 运行

```bash
./printer-arm64 -config ./config.json
```

## 常用环境变量覆盖

```bash
export MQTT_BROKER=mqtt.serbia70.com
export MQTT_PORT=443
export MQTT_USERNAME=CHANGE_ME
export MQTT_PASSWORD=CHANGE_ME
export SHOP_SLUG=CHANGE_ME
export CONFIG_SYNC_BASE_URL=https://food2api.serbia70.com
export PRINTER_VENDOR=1155
export PRINTER_PRODUCT=1803
export MAX_RETRIES=10
./printer-arm64 -config ./config.json
```

## 仓库清理约定

- 不提交 `config.json`
- 不提交 `printer-arm64` / `printer-n1.exe` 等编译产物
- 不保留嵌套 `.git/` 目录
