tag=$(date +%Y-%m-%d-%s)
echo "using tag $tag"
docker build . -t daniel98/dron8s:$tag
docker push daniel98/dron8s:$tag