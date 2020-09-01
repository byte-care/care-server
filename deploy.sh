#!/bin/bash
wget -q https://webplus-cn-shenzhen.oss-cn-shenzhen.aliyuncs.com/cli/wpctl-linux
mv ./wpctl-linux ./wpctl
chmod +x ./wpctl

./wpctl configure --access-key-id "$ALICLOUD_ACCESS_KEY" --access-key-secret "$ALICLOUD_SECRET_KEY" --region "$ALICLOUD_REGION"

go build .
zip upload.zip care-server

./wpctl env:deploy upload.zip --app Kan --env KanEnv