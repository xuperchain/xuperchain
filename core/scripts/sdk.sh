docker build -t xuperchain:sdk  -f Dockerfile.ct . --no-cache
docker ps -f 'name=xchain' &&  docker stop xchain
docker run --cap-add=SYS_PTRACE --security-opt seccomp=unconfined  -it --name xchain --rm  -d   -p 1234:1234  xuperchain:sdk  ./xchain
docker exec -it xchain bash  scripts/start.sh || docker exec -it xchain bash
#docker exec -it xchain bash