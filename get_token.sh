image=library/ubuntu
tag=18.04
token=$(curl --silent "https://auth.docker.io/token?scope=repository:$image:pull&service=registry.docker.io" | jq -r '.token')
curl --header "Accept: application/vnd.docker.distribution.manifest.v2+json" --header "Authorization: Bearer $token" "https://registry-1.docker.io/v2/$image/manifests/$tag" | jq -r '.config.digest'

image=library/ubuntu
tag=latest
#token=$(curl --silent "https://auth.docker.io/token?scope=repository:$image:pull&service=registry.docker.io" | jq -r '.token')
curl --header "Accept: application/vnd.docker.distribution.manifest.v2+json" --header "Authorization: Bearer $token" "https://registry-1.docker.io/v2/$image/manifests/$tag" | jq -r '.config.digest'
