#!/usr/bin/env bash
set -euo pipefail

make docker-image IMAGE_PLATFORMS=linux/amd64 DOCKER_TAG=janek/navidrome:latest
docker save janek/navidrome:latest | gzip > navidrome-custom.tar.gz

rsync -rlDvz --delete --no-o --no-g --no-perms -e "ssh -p 10331" navidrome-custom.tar.gz janek@tadek331.mikrus.xyz:/home/janek/

# ssh -p 10331 janek@tadek331.mikrus.xyz << 'EOF'
#   set -euo pipefail
#   docker load < /home/janek/navidrome-custom.tar.gz
#   cd /opt/navidrome
#   docker compose down
#   docker compose up -d
# EOF
